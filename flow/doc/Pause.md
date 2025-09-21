# Pause generation using interrupts

_Interrupts_ are a special kind of [tool](/docs/tool-calling) that can pause the
LLM generation-and-tool-calling loop to return control back to you. When
you're ready, you can then _resume_ generation by sending _replies_ that the LLM
processes for further generation.

The most common uses for interrupts fall into a few categories:

- **Human-in-the-Loop:** Enabling the user of an interactive AI
  to clarify needed information or confirm the LLM's action
  before it is completed, providing a measure of safety and confidence.
- **Async Processing:** Starting an asynchronous task that can only be
  completed out-of-band, such as sending an approval notification to
  a human reviewer or kicking off a long-running background process.
- **Exit from an Autonomous Task:** Providing the model a way
  to mark a task as complete, in a workflow that might iterate through
  a long series of tool calls.

## Before you begin

All of the examples documented here assume that you have already set up a
project with Genkit dependencies installed. If you want to run the code
examples on this page, first complete the steps in the
[Get started](/docs/get-started/) guide.

Before diving too deeply, you should also be familiar with the following
concepts:

- [Generating content](/docs/models/) with AI models.
- Genkit's system for [defining input and output schemas](/docs/flows/).
- General methods of [tool-calling](/docs/tool-calling/).

## Overview of interrupts

At a high level, this is what an interrupt looks like when
interacting with an LLM:

1.  The calling application prompts the LLM with a request. The prompt includes
    a list of tools, including at least one for an interrupt that the LLM
    can use to generate a response.
2.  The LLM generates either a complete response or a tool call request
    in a specific format. To the LLM, an interrupt call looks like any
    other tool call.
3.  If the LLM calls an interrupting tool,
    the Genkit library automatically pauses generation rather than immediately
    passing responses back to the model for additional processing.
4.  The developer checks whether an interrupt call is made, and performs whatever
    task is needed to collect the information needed for the interrupt response.
5.  The developer resumes generation by passing an interrupt response to the
    model. This action triggers a return to Step 2.

## Define manual-response interrupts

The most common kind of interrupt allows the LLM to request clarification from
the user, for example by asking a multiple-choice question.

For this use case, use the `genkit.DefineTool()` function and call `ctx.Interrupt()`:

```go
import (
	"context"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type QuestionInput struct {
	Choices    []string `json:"choices" jsonschema:"description=the choices to display to the user"`
	AllowOther bool     `json:"allowOther,omitempty" jsonschema:"description=when true, allow write-ins"`
}

func main() {
	ctx := context.Background()

	g := genkit.Init(ctx)

	askQuestion := genkit.DefineTool(
		g,
		"askQuestion",
		"use this to ask the user a clarifying question",
		func(ctx *ai.ToolContext, input QuestionInput) (string, error) {
			return "", ctx.Interrupt(&ai.InterruptOptions{
				Metadata: map[string]any{
					"question": input,
				},
			})
		},
	)
}
```

Note that the output type of an interrupt tool corresponds to the response data
you will provide as opposed to something that will be automatically populated
by a tool function.

### Use interrupts

Interrupts are passed into the `WithTools()` option when generating content, just like
other types of tools. You can pass both normal tools and interrupts to the
same `Generate` call:

```go
response, err := ai.Generate(ctx, g,
	ai.WithPrompt("Ask me a movie trivia question."),
	ai.WithTools(askQuestion),
)
if err != nil {
	panic(err)
}
```

Genkit immediately returns a response on receipt of an interrupt tool call.

### Respond to interrupts

If you've passed one or more interrupts to your generate call, you
need to check the response for interrupts so that you can handle them:

```go
// You can check the 'FinishReason' attribute of the response
if response.FinishReason == "interrupted" {
	fmt.Println("Generation interrupted.")
}
// or you can check to see if any interrupt requests are on the response
interrupts := response.Interrupts()
if len(interrupts) > 0 {
	fmt.Printf("Found %d interrupts\n", len(interrupts))
}
```

Responding to an interrupt is done using the `ai.WithToolResponses()` option on a subsequent
`Generate` call, making sure to pass in the existing message history. You can use the tool's
`Respond` method to help construct the response.

Once resumed, the model re-enters the generation loop, including tool
execution, until either it completes or another interrupt is triggered:

```go
response, err := ai.Generate(ctx, g,
	ai.WithPrompt("Help me plan a backyard BBQ."),
	ai.WithSystemPrompt("Ask clarifying questions until you have a complete solution."),
	ai.WithTools(askQuestion),
)
if err != nil {
	panic(err)
}

for response.FinishReason == "interrupted" {
	var answers []*ai.Part
	// multiple interrupts can be called at once, so we handle them all
	for _, part := range response.Interrupts() {
		// use the `Respond` method on our tool to populate answers
		answers = append(answers, askQuestion.Respond(part, askUser(part.ToolRequest.Input), nil))
	}

	response, err = ai.Generate(ctx, g,
		ai.WithMessages(response.History()...),
		ai.WithTools(askQuestion),
		ai.WithToolResponses(answers...),
	)
	if err != nil {
		panic(err)
	}
}

// no more interrupts, we can see the final response
fmt.Println(response.Text())
```

## Tools with restartable interrupts

Another common pattern for interrupts is the need to _confirm_ an action that
the LLM suggests before actually performing it. For example, a payments app
might want the user to confirm certain kinds of transfers.

For this use case, you can use the standard `DefineTool` method to add custom
logic around when to trigger an interrupt, and what to do when an interrupt is
_restarted_ with additional metadata.

### Define a restartable tool

Every tool has access to special helpers in the `ToolContext`:

- `Interrupt`: when called, this method returns a special kind of error that
  is caught to pause the generation loop. You can provide additional metadata
  as an object.
- `Resumed`: when a request from an interrupted generation is restarted using
  the `WithToolRestarts()` option (see below), this field contains the
  metadata provided when restarting.

If you were building a payments app, for example, you might want to confirm with
the user before making a transfer exceeding a certain amount:

```go
type TransferInput struct {
	ToAccountID string `json:"toAccountId" jsonschema:"description=the account id of the transfer destination"`
	Amount      int    `json:"amount" jsonschema:"description=the amount in integer cents (100 = $1.00)"`
}

type TransferOutput struct {
	Status  string `json:"status" jsonschema:"description=the outcome of the transfer"`
	Message string `json:"message,omitempty"`
}

transferMoney := genkit.DefineTool(
	g,
	"transferMoney",
	"Transfers money between accounts.",
	func(ctx *ai.ToolContext, input TransferInput) (TransferOutput, error) {
		// if the user rejected the transaction
		if ctx.Resumed != nil {
			if status, ok := ctx.Resumed["status"].(string); ok && status == "REJECTED" {
				return TransferOutput{
					Status:  "REJECTED",
					Message: "The user rejected the transaction.",
				}, nil
			}
		}

		// trigger an interrupt to confirm if amount > $100
		if ctx.Resumed == nil || (ctx.Resumed != nil && ctx.Resumed["status"] != "APPROVED") {
			if input.Amount > 10000 {
				return TransferOutput{}, ctx.Interrupt(&ai.InterruptOptions{
					Metadata: map[string]any{
						"message": "Please confirm sending an amount > $100.",
					},
				})
			}
		}

		// Complete the transaction if not interrupted
		return doTransfer(input), nil
	},
)
```

In this example, on first execution (when `Resumed` is nil), the tool
checks to see if the amount exceeds $100, and triggers an interrupt if so. On
second execution, it looks for a status in the new metadata provided and
performs the transfer or returns a rejection response, depending on whether it
is approved or rejected.

### Restart tools after interruption

Interrupt tools give you full control over:

1. When an initial tool request should trigger an interrupt.
2. When and whether to resume the generation loop.
3. What additional information to provide to the tool when resuming.

In the example shown in the previous section, the application might ask the user
to confirm the interrupted request to make sure the transfer amount is okay:

```go
response, err := ai.Generate(ctx, g,
	ai.WithPrompt("Transfer $1000 to account ABC123"),
	ai.WithTools(transferMoney),
)
if err != nil {
	panic(err)
}

for response.FinishReason == "interrupted" {
	var confirmations []*ai.Part
	// multiple interrupts can be called at once, so we handle them all
	for _, part := range response.Interrupts() {
		confirmations = append(confirmations,
			// use the 'Restart' method on our tool to provide `Resumed` metadata
			transferMoney.Restart(
				part,
				&ai.RestartOptions{
					// send the tool request input to the user to respond. assume that this
					// returns `{status: "APPROVED"}` or `{status: "REJECTED"}`
					ResumedMetadata: requestConfirmation(part.ToolRequest.Input),
				},
			),
		)
	}

	response, err = ai.Generate(ctx, g,
		ai.WithMessages(response.History()...),
		ai.WithTools(transferMoney),
		ai.WithToolRestarts(confirmations...),
	)
	if err != nil {
		panic(err)
	}
}

// no more interrupts, we can see the final response
fmt.Println(response.Text())
```

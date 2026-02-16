# Development process

1. Understand the problem and requirements
2. Define assumptions and constraints
3. Explore architectural approaches
4. Choose the best approach based on trade-offs
5. Implement the solution
6. Test the solution
7. Improve ...

---

## Steps 1-4

I just wrote the `@docs/001-thought-process.md` document by myself then discuss with Perplexity AI to improve the trade-offs.

```prompt
As a Systems Architect, let's discuss the project details below to understand more about the trade-offs and improve the document result:

@001-thought-process.md
```

Result: Perplexity AI helps me with "Asynchronous Cons" and the "Aspect" table.

---

> I decided to ask for sequence diagrams for the approaches them to `@docs/002-sync-approach-chart.md` and `@docs/003-async-approach-chart.md`.

```prompt
Generate the sequence charts with MermaidJS for the sync and async approaches.
```

---

Create the project constitution by using SpecKit.

```prompt
/speckit.constitution Create the constitution document based on docs @docs/001-thought-process.md and @docs/002-project-architecture.md
```

## Step 5

Create the specification.

```prompt
/speckit.specify The API client should be able to send an open-format address to the API and receive the normalized address. Only US address required.
```

Output: The specification document @specs/001-address-normalization/spec.md.

Create the plan.

```prompt
/speckit.plan Plan the technical implementation that covers all the specifications created on 001-address-normalization. The endpoint to be used is "/api/v1/validate-address", following by ValidateRequest and ValidateResponse. To implement the parse and normalization process use the "gopostal" library as a dependency to ValidateAddressUsecase.
```

Output: The plan document @specs/001-address-normalization/plan.md.

Create the tasks.

```prompt
/speckit.tasks
```

Implement the tasks.

```prompt
/speckit.implement
```

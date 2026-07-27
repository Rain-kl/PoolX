# application

Application services orchestrate domain use cases.

- Keep packages thin: validate input, call repositories/domain helpers, return domain errors.
- Copy `example/` when adding a new vertical slice.
- Do not put HTTP binding or SQL here.

# Event Checklist

## Before editing

- confirm bounded context
- confirm target file name
- confirm whether payload schema already exists

## While editing

- keep version explicit
- keep producer and trigger explicit
- keep status explicit
- keep envelope fields explicit
- prefer schema references for reusable payloads

## Final review

- file lives under `contracts/events/v1/`
- filename follows lowercase snake_case plus version suffix
- semantic meaning is stable and explicit
- payload schema references are correct when used
- no app code is being treated as a parallel event source
- changes fit generated TS and Python consumption


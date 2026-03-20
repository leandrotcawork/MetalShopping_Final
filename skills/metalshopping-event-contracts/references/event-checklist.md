# Event Contract Checklist

## Before editing
- [ ] Bounded context confirmed
- [ ] Target filename confirmed: `<domain>_<event>.v1.json`
- [ ] Payload schema checked in `contracts/api/jsonschema/`

## While editing
- [ ] Version explicit in filename and metadata
- [ ] Producer and trigger explicit
- [ ] Status explicit (proposed | active | deprecated)
- [ ] Schema reference used for reusable payloads

## Final
- [ ] File in `contracts/events/v1/`
- [ ] Semantic meaning stable and explicit
- [ ] No app code treated as parallel event source
- [ ] Idempotency key pattern documented: `"event_name:aggregate_id"`

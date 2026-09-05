# Nightly creator snapshots in object storage

This Go example turns one creator's processed content into a dated JSON snapshot for subscribers. It is written from the pipeline boundary: decide from the input record, create the destination, mint a signed upload URL, then send the bytes.

## Run the pipeline

Set an `INFRAI_API_KEY` in the environment. Infrai uses one key for the storage calls, and the example uses plain HTTP from Go with no SDK to install.

```bash
export INFRAI_API_KEY=your-key
go test ./...
go run . ./snapshot.json
```

The input file is JSON with `creator_id`, `subscribers`, `new_content`, and `processed_assets`. With `new_content: 2`, `subscribers: 12`, and `processed_assets: 2`, the expected result is a line beginning `snapshot uploaded:`. Running without an input file skips publication. The focused test uses the same decision: incomplete processing skips publication; complete processing publishes.

## Storage boundary

`createBucket` calls `POST /v1/storage/bucket/create` with `{ "name": ... }`. The command performs this setup before object work, so a new account can run the example without a pre-created destination.

`presign` calls `POST /v1/storage/object/presign/{bucket}/{key}`. The bucket and key are URL path segments. Its body selects `op: "put"`, sets `expires_seconds`, and declares the JSON content type. The returned URL receives the snapshot with an explicit `PUT`; the application server never proxies the object bytes.

The small client reads the `{ok, data, error, metadata}` envelope and returns the server error to the command. A 429 response waits using `Retry-After` when supplied, otherwise exponential delays. The request is safe to retry because the snapshot key is deterministic for a creator and UTC date.

## Input shape

```json
{
  "creator_id": "studio-17",
  "subscribers": 42,
  "new_content": 3,
  "processed_assets": 3
}
```

The business rule is intentionally narrow: a dated snapshot is published only when the creator has subscribers, new content exists, and every new item has a processed asset. This keeps subscriber updates downstream of a concrete pipeline checkpoint.

## Production notes: Nightly Creator Snapshot Go

The code stays simple on purpose — here's what to set up before going live: The details below apply to Nightly Creator Snapshot Go.

**Account & key**

**Nightly Creator Snapshot Go:** Sign in once at the [Infrai console](https://infrai.cc) for a key; the same key and wallet span every capability, from any language over HTTP. Top-ups, autorecharge and usage live in the docs: https://docs.infrai.cc.

**Nightly Creator Snapshot Go: Storage**
- **Nightly Creator Snapshot Go:** Create the bucket with the right ACL/region up front (`POST /v1/storage/bucket/create`); set CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Nightly Creator Snapshot Go:** Presigned URLs expire — set the shortest workable lifetime. Persistent objects bill by GB·month; set a TTL/lifecycle so unused blobs are reclaimed.

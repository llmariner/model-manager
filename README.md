# model-manager

## Running Locally

```bash
make build-server
./bin/server run --config config.yaml
```

`config.yaml` has the following content:

```yaml
httpPort: 8080
grpcPort: 8081
internalGrpcPort: 8082

debug:
  standalone: true
  sqlitePath: /tmp/model_manager.db
```

You can then connect to the DB.

```bash
sqlite3 /tmp/model_manager.db
# Run the query inside the database.
insert into models
  (model_id, tenant_id, created_at, updated_at)
values
  ('my-model', 'fake-tenant-id', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
```

You can then hit the endpoint.

```bash
curl http://localhost:8080/v1/models

grpcurl -d '{"base_model": "base", "suffix": "suffix", "tenant_id": "fake-tenant-id"}' -plaintext localhost:8082 list llmoperator.models.server.v1.ModelsInternalService/CreateModel
```

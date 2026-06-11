# Schema reference (NOT auto-loaded)

These hand-written SQL files document the demo schemas and how the Lake-Base
(`rc_*`) tables are structured. They are kept for readability only.

**They are not executed on MariaDB startup.** The entrypoint ignores
subdirectories of `/docker-entrypoint-initdb.d`, so files in this `reference/`
folder never run as DB init.

> Note: `01_init_lakebase.sql` is still a build dependency — it is the single
> source of truth for the `rc_*` schema and is embedded into the backend image
> (`//go:embed`, see `backend/internal/lakebase/migrate.go` and
> `deploy/Dockerfile.backend`) for auto-migration. Do not delete it.

The actual cold-start seed is a single file one level up:

    deploy/init/mariadb/01_atlas_demo.sql.gz

It is a full logical dump (schema + data + Rich Context + 2048-dim vector
embeddings) of the reference demo database. Loading it gives a brand-new
machine the exact working demo — all five datasources, their Rich Context, and
working vector retrieval — **without needing any embedding/LLM API key**.

To regenerate the seed from a running stack:

```bash
docker exec lucid-mariadb sh -c \
  "mariadb-dump -uroot -p\"$MARIADB_ROOT_PASSWORD\" \
   --databases lucid lucid_evolution spider_tvshow spider_flight spider_wta tpch_enterprise \
   --hex-blob --routines --single-transaction --skip-comments --no-tablespaces" \
  | { cat; printf '\nGRANT ALL PRIVILEGES ON \x60lucid\x60.* TO '"'"'lucid'"'"'@'"'"'%%'"'"';\n'; } \
  | gzip -9 > deploy/init/mariadb/01_atlas_demo.sql.gz
```

(Grant the `lucid` app user on every demo database; the entrypoint only grants
on `lucid` itself. See the trailing `GRANT` lines in the generated file.)

# Aurora PostgreSQL Cluster

This Acorn creates an Aurora PostgreSQL cluster running on AWS RDS.

## Usage

From the CLI you can run the following command to create an Aurora PostgreSQL cluster

```shell
acorn run -n rds-postgresql-cluster ghcr.io/acorn-io/aws/rds/aurora/postgresql/cluster:v0.#.#
```

From an Acornfile you can create the cluster by using the PostgreSQL acorn too
```cue
services: pg: {
    image: "ghcr.io/acorn-io/aws/rds/aurora/postgresql/cluster:v0.#.#"
}
containers: app: {
    build: context: "./"
    ports: publish: ["8080/http"]
    env: {
      PGDATABASE: "@{service.pg.data.dbName}"
      PGHOST: "@{service.pg.data.address}"
      PGPORT: "@{service.pg.data.port}"
      PGUSER: "@{service.pg.secrets.admin.username}"
      PGPASSWORD: "@{service.pg.secrets.admin.password}"
    }
}
```

To run from source you can `acorn run .` in this directory.

## Arguments

| Name                      | Description                                                                                                                                             | Type   | Default   |
|---------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|--------|-----------|
| adminUsername             | Name of the root/admin user.                                                                                                                            | string | postgres  |
| username                  | Name of an additional user to create. This user will have complete access to the database. If left blank, no additional user will be created.           | string |           |
| dbName                    | Name of the default database.                                                                                                                           | string | postgres  |
| deletionProtection        | Must be set to false in order for the RDS db to be deleted.                                                                                             | bool   | false     |
| instanceClass             | The instance class the database server will use. Options are: burstable, burstableGraviton, memoryOptimized. Updating this setting will cause downtime. | string | burstable | 
| instanceSize              | The size of the instance to use. Options are: medium, large, xlarge, or 2xlarge. Updating this setting will cause downtime.                             | string | medium    |
| parameters                | RDS PostgreSQL database parameters to apply to the cluster. Must be key-value string pairs (ex. max_connections: "1000").                               | object | {}        |
| skipSnapshotOnDelete      | Do not take a final snapshot on delete or update and replace operations. If enabled the DB will gone forever if deleted or replaced.                    | bool   | false     |
| enablePerformanceInsights | Enables performance insights when true.                                                                                                                 | bool   | false     |
| tags                      | Key value pairs of tags to apply to the RDS cluster and all other resources.                                                                            | object | {}        |

## Output Services

```cue
services: rds: {
  default: true
  address: "${ADDRESS}"
  ports: [${PORT}]
  secrets: ["admin"]
  data: {
    address: "${ADDRESS}"
    port: "${PORT}"
    dbName: "${DB_NAME}"
  }
}

secrets: "admin": {
 type: "basic"
 data: {
    username: "${ADMIN_USERNAME}"
    password: "${ADMIN_PASSWORD}"
 }
}

// If username is set, create a user secret
secrets: user: {
    type: "basic"
    data: {
        username: "${USERNAME}"
        password: "${PASSWORD}"
    }
}
```

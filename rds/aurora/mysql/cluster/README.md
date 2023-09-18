# Aurora MySQL Cluster

This Acorn creates an Aurora MySQL cluster running on AWS RDS service. This Acorn is best used when you have a steady state workload that can be serviced by a VM instance. If you have a burstable workload you could consider using the Aurora Serverless v2 Acorn.

## Usage

From the CLI you can run the following command to create an Aurora MySQL cluster:

```shell
acorn run -n rds-mysql-cluster ghcr.io/acorn-io/aws/rds/aurora/mysql/cluster:v1.#.#
```

From an Acornfile you can use the following Acorn:

```cue
services: "rds-mysql-cluster": {
    image: ghcr.io/acorn-io/aws/rds/aurora/mysql/cluster:v1.#.#
}

containers: wp: {
    image: wordpress:latest
    ports: publish: "80/http"
    env: {
        WORDPRESS_DB_HOST: "@{services.rds-mysql-cluster.address}"
        WORDPRESS_DB_USER: "@{services.rds-mysql-cluster.secrets.amdin.username}"
        WORDPRESS_DB_PASSWORD: "@{services.rds-mysql-cluster.secrets.admin.password}"
        WORDPRESS_DB_NAME: "instance"
    }
}
```

## Arguments

| Name | Description | Type |
|------|-------------|------|
| adminUsername | Name of the root/admin user. Default is admin. | string |
| username | Name of an additional user to create. This user will have complete access to the database. If left blank, no additional user will be created. | string |
| dbName | Name of the database instance. Default is instance. | string |
| deletionProtection | Deletion protection, you must set to false in order for the RDS db to be deleted. Default is false. | bool |
| instanceClass | The instance class for the database server to use. Default is "burstable".  - burstable (good for dev/test and light workloads.) - burstableGraviton (good for dev/test and light workloads.) - memoryOptimized (good for memory intensive workloads.) **Updating this setting will cause downtime on the RDS instance.** | string |
| instanceSize | The instance size to use.(medium, large, xlarge, or 2xlarge) Default is "medium". Not all instance sizes are available in all regions. **Updating this setting will cause downtime on the RDS instance.** | string |
| parameters | RDS MySQL Database Parameters to apply to the cluster. Must be k/v string pairs(ex. max_connections: "1000"). | object |
| skipSnapshotOnDelete | Do not take a final snapshot on delete or update and replace operations. Default is false. If skip is enabled the DB will be gone forever if deleted or replaced. | bool |
| enablePerformanceInsights | Enable Performance insights. Default is false. | bool |
| tags | Key value pairs of tags to apply to the RDS cluster and all other resources. | object |

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

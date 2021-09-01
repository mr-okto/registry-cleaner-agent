# Registry Cleaner Agent



## Docker Registry Agent

Acts as a proxy for Registry API.

Automates garbage collection inside Docker Registry container.

Garbage collector restarts Registry container in maintenance mode and removes unused blob files (file system layers no longer required).

Registry is available in read-only mode during garbage collection.

Healthcheck tests availability of registry API. 

Registry Spec:

https://docs.docker.com/registry/spec/api/

Additional routes:

`GET /v2/status` - healthcheck  
`GET /v2/garbage` - index garbage blobs    
`GET /v2/<name>/manifests/<tag>/digest` - get image digest   
`DELETE /v2/<name>/manifests/<digest>` - remove image manifest   
`DELETE /v2/garbage` - run garbage collector  


Garbage removal is launched automatically using CRON schedule (config/agent.toml).

 

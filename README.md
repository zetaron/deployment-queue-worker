# Deployment Queue Worker
Receives deployment payloads via a redis queue, which might be submitted by the [zetaron/github-hook-receiver](https://github.com/zetaron/github-hook-receiver).

## Usage
With `docker-compose` (recomended for development):
```shell
docker-compose up -d && docker-compose logs -f worker
```

With `docker swarm` (recomended for production):
```shell
./hooks/deploy
```

## Configuration
- **DEPLOYMENT_ID_TEMPLATE** [default="deployment-{{.deployment.id}}"]
- **CACHE_VOLUME_NAME_TEMPLATE** [default="{{.repository.name}}-{{.deployment.environment}}-deployment-{{.deployment.id}}"]
- **SECRETS_VOLUME_NAME_TEMPLATE** [default="{{.repository.name}}-{{.deployment.environment}}-secrets"]
- **REDIS_URL** [default=redis:6379]
- **REDIS_DATABASE** [default=1]
- **WORKER_COUNT** [default=1]

## Integration
To enable your project to be deployed by the `deployment-queue-worker` you have to create a `docker-compose.deploy.yml` and some scripts inside `hooks/`.

### docker-compose.deploy.yml
```yaml
version: '2'

services:
  deploy:
    image: zetaron/github-deployment-worker:1.0.0
    volumes:
      - secrets:/var/cache/secrets
      - cache:/var/cache/deployment

volumes:
  secrets:
    external:
      name: $SECRETS_VOLUME_NAME
  cache:
    external:
      name: $DEPLOYMENT_CACHE_VOLUME_NAME
```

> *Note:* The `$SECRETS_VOLUME_NAME` and `$DEPLOYMENT_CACHE_VOLUME_NAME` are provided by the queue-worker.
> If you need to start a docker process wich has access to either of these volumes just extend your service to carry the two via the `environment` section.

### hooks/
Depending on which deployment worker you use you will be able to use different hooks.
There are three core hooks each deployment worker should implement:
- pre_deploy
- deploy
- post_deploy

If you do not provide a deploy hook the worker will fail, as it doesn't know how to deploy your project.

## List of deployment-worker
> *Note:* If you got something to share just create a Pull Request and add your worker to the list below.
> The schema is: docker image name -> link to git repository

- [zetaron/github-deployment-worker](https://github.com/zetaron/github-deployment-worker)
- [zetaron/hook-deployment-worker](https://github.com/zetaron/hook-deployment-worker)
- [zetaron/docker-deployment-worker](https://github.com/zetaron/docker-deployment-worker)

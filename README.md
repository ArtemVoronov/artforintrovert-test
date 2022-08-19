# How to build and run
`docker-compose build && docker-compose up`

# Configuration
Add appropriate `.env` file in case of different environment. Default settings (if there is no `.env` config) is the following:
```
# common settings
APP_PORT=3000
CORS='*'
APP_MODE=debug # or release

# db settings
DATABASE_USERNAME=mongo_admin
DATABASE_PASSWORD=mongo_admin_password
DATABASE_NAME=testdb
DATABASE_HOST=mongo
DATABASE_PORT=27017
DATABASE_CONNECT_TIMEOUT_IN_SECONDS=30
DATABASE_QUERY_TIMEOUT_IN_SECONDS=30

# cache settings
UPDATE_CACHE_MIN_INTERVAL_IN_SECONDS=30
UPDATE_CACHE_MAX_INTERVAL_IN_SECONDS=86400 # 24 hours
UPDATE_CACHE_INTERVAL_FACTOR=2 # increasing twice in case of error
```

# API endpoints

## Entities
Any request with body could have a json object with only two attrs:
- ```id``` (optional fo update)
- ```data``` (optional for delete)

Example
```
{
    "id": "62ffcac20074ec24bbb5810d",
    "data": "exponent"
}
```

## Example 1 (get all)
Request

```GET http://localhost:3000/api/v1/records/```

Response
```
[
    {
        "id": "62ffcac20074ec24bbb5810d",
        "data": "pi"
    },
    {
        "id": "62ffcac90074ec24bbb5810e",
        "data": "exponent"
    }
]
```

## Example 2 (create)
Request 

```
PUT http://localhost:3000/api/v1/records/
{
    "data": "exponent"
}
```

Response
```
"62ffcac90074ec24bbb5810e"
```

## Example 3 (update)
Request

```
PUT http://localhost:3000/api/v1/records/
{
    "id": "62ffcac90074ec24bbb5810e",
    "data": "planck"
}
```

Response
```
"Done"
```


## Example 4 (delete)
Request

```
DELETE http://localhost:3000/api/v1/records/
{
    "id": "62ffcac90074ec24bbb5810e"
}
```

Response
```
"Done"
```
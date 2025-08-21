# Auth Service

## Auth Basic
* Project located on `/auth-basic`
* Run on port 8080
* Expose `/login` endpoint, for authentication using parameter: `username` and `password`
* Password hash using bcrypt
* login on success return jwt
* Run via 'make auth-basic'

## Auth Improved
### Http server
* Project located on `/auth-improved`
* Run on port 8081
* Expose `/login` endpoint, similar with auth-basic
* Password hash using bcrypt
* login on success return jwt
* Run via 'make auth-improved'
* Has feature toggle, on-off to load 'user' data from cache (redis) or load from db, when off, load from db
* When on, it will load the data from redis (precomputed by precache-worker)

### User Precache worker
* Project located on `/precache-worker`
* Run periodically each minute, via cron (robfig/cron)
* Has feature toggle, on-off. If off, it just skip the process
* When on, it will load all user in database, per X rows (batch size, configurable)
* Run via 'make precache-worker'
* Cache key is 'username'


## User seeder
* Run via 'make seeder <N>', where <N> is number of user to be generated
* It will generate sequential username with format <8-digit-zero-padded>@katakode.com, insert the data to db as batch on N per looping, all with same password (simulation), if it already exist, skip, no need to error. for N=100, then 100 sequential users like '00000001@katakode.com', '00000002@katakode.com', etc. will be generated.

## Load tester
* use k6 javascript, it will hit '/login' 
* create two scenario (files), one for 'auth-basic', one for 'auth-improved'
* run for 5 minute, with vu of 10 users

# Infra
## Docker compose
* redis (max memory 256MB)
* mysql (max memory 500MB)

# Tech stack
* go 1.24
* sqlx for db abstraction
* logging use slog
* docker-compose for docker
* godotenv for env loader
* no need for migration, just create sql file for schema, I'll run it manually to db
* user table should have index for 'username'
* https://github.com/go-jose/go-jose for jwt library
* makefile to run the commands above, make it simple
* Git, with .gitignore for stuff

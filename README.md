# URL Shortener DEMO using GO

> Database: MongoDB
> Use Recaptcha v3 for robot checking
> Use gRPC for services internal communication
> Use Docker to launch microservices

### ToDo
- [x] Build prototype
- [x] Recaptcha v3
- [] Shorten URL algorithm
- [x] Docker settings
- [] gRPC communication

### Addition
- [] Switch DB to Redis



### steps:

build urlshortener docker image:
docker build -t urlshortener:1.0 . -f Dockerfile

pull mongodb docker image:
docker pull mongo

run mongodb on docker container:
docker run -p 27017:27017 -v $PWD/db:/data/db -d --name mongo_container mongo

run urlshortener on docker container:
docker run -p 8080:8080 -it --link mongo_container:mongo urlshortener:1.0 <reCaptcha public key> <reCaptcha private key>

# urlShortener



steps:

build urlshortener docker image:
docker build -t urlshortener:1.0 . -f Dockerfile

pull mongodb docker image:
docker pull mongo

run mongodb on docker container:
docker run -p 27017:27017 -v $PWD/db:/data/db -d --name mongo_container mongo

run urlshortener on docker container:
docker run -p 8080:8080 -it --link mongo_container:mongo urlshortener:1.0 <reCaptcha public key> <reCaptcha private key>

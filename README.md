<h1 align="center">📡 MCScan</h1>
<h4 align="center">Minecraft Server scanner written in Go</h4>

## Inspiration
This project is highly inspired by the [LiveOverflow Minecraft hacking series](https://www.youtube.com/watch?v=VIy_YbfAKqo).  
This tool is used to scan the entire Internet on port 25565 and send a Minecraft ping packet to test if the server candidate is a Minecraft Server.
On success the metadata is saved in a simple NoSQL database.

## Setup
A working docker container is published on Docker Hub.
Simply use the `docker-compose.yml` to start MongoDB and start scanning.  
MCScan restarts the container automatically after a whole scan and begins from start.  
Good to know: MCScan has implemented the default exclude.conf by masscan to ignore unrouted networks.

### Warning
⚠ The MongoDB Port 27017 is exposed by default. Please use a more secure way to access your MongoDB.

## Explore data
To explore the found server use [Mongo Explorer](https://hub.docker.com/_/mongo-express) as a webtool or [Mongo Compass](https://www.mongodb.com/de-de/products/compass) for more complex query.
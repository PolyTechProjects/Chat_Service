# Chat Service
## _Project as a task for Polytech course_

![Build Status](https://github.com/PolyTechProjects/Chat_Service/actions/workflows/go.yml/badge.svg)

### How to start

Launch infra_start.sh to start databases, redis, file storage.

```sh
./infra_start.sh
```

Then launch services_start.sh to start applications.

```sh
./services_start.sh
```

In each directory there is script 'build.sh' which you can launch to build new version of particular service. If you made changes in multiple services you may launch '[build_all.sh]' script that is in main directory. '[build_all.sh]' will build binary file and docker image. 'build.sh' will only build binary file.

'[open_websocket_connection.sh]' is script that takes 3 necessary parameters such as user's login, password and userId. This will login user using his login and password, take from response headers tokens and open websocket session using '[websocat]'. This script makes testing of chat logic much easier.

[build.sh]: <https://github.com/joemccann/dillinger>
[build_all.sh]: <https://github.com/PolyTechProjects/Chat_Service/blob/main/build_all.sh>
[open_websocket_connection.sh]: <http://daringfireball.net>
[websocat]: <https://github.com/vi/websocat>


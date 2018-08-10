#!/bin/bash
docker build -t godwhoa/sandbox ..
# docker run --rm --cap-drop=ALL -v $(pwd)/../main.c:/src/main.c:ro --net none --memory 256m --memory-swap 320m --stop-timeout 60 --pids-limit 512 godwhoa/sandbox
docker run --rm --security-opt="no-new-privileges:true,seccomp:../profiles/profile.json" --cap-drop=ALL -v $(pwd)/../testfiles/main.c:/src/main.c:ro --net none --memory 256m --memory-swap 320m --stop-timeout 60 --pids-limit 512 godwhoa/sandbox
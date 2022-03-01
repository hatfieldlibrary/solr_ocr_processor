# Configuration File

When running the container, the volume mount point for the configuration file is `/app/configs`. 

`docker run -p 3000:3000 -v /absolute/path/to/config/directory:app/configs`

Note that you must also provide a volume for logging. So the full command is:

`docker run -p 3000:3000 -v /absolute/path/to/config/directory:app/configs -v /absolute/path/to/log/directory:app/logs`

If you want to bypass the volume during development you can change the paths in
main.go. Then edit the configuration file in source (app/configs/config.yml)


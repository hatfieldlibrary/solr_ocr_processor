# Configuration File

When deploying the container you must add a configuration file to the host file system.
That file must be based on the example provided here.

When running the container, the volume mount point for the configuration file is `/app/configs`. 

`docker run -p 3000:3000 -v /absolute/path/to/config/directory:app/configs`

Note that you must also provide a volume for logging. So the full command is:

`docker run -p 3000:3000 -v /absolute/path/to/config/directory:app/configs -v /absolute/path/to/log/directory:app/logs`


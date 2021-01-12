# Change log

## 2021-01-12

- Expansion of `.` and `~` in volume configurations is now also supported if the config is all on a single line. (e.g.: `-v=~/exchange`). Close #1
- Expansion of `.` and `~` in volume configurations is now also supported if the config is espressed with the `--mount` format (see <https://docs.docker.com/storage/bind-mounts/>)

## 2021-01-11 

- Modified sample `dockerstarter.yaml`
- Added (optional) configuration `image` to a container configuration to specify the image name
- Added facilities to check if an image exists and to `docker pull` it if missing

# Build
```shell
go build -o tasker ./cmd/app/main.go
sudo mv tasker /usr/local/bin/tasker
```

# TODO
- Use Cobra: https://github.com/spf13/cobra

# MongoDB
## Run
docker run -it --rm -p 27017:27017 -v /home/egorka/Downloads/test_mongo:/data/db mongo:5.0.6

# Command
tracker [command]
statistic - Show statistic
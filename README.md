# Build
```shell
go build -o tracker ./cmd/app/main.go
sudo mv tracker /usr/local/bin/tracker
```

# TODO
- Use Cobra: https://github.com/spf13/cobra

# MongoDB
## Run
docker run -it --rm -p 27017:27017 -v /home/egorka/Downloads/test_mongo:/data/db mongo:5.0.6

# Command
tracker [command]
statistic - Show statistic
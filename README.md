# tg-reminder

<div align="center">

[![build](https://github.com/mezk/tg-reminder/actions/workflows/ci.yml/badge.svg)](https://github.com/mezk/tg-reminder/actions/workflows/ci.yml)&nbsp;[![Coverage Status](https://coveralls.io/repos/github/MEZk/tg-reminder/badge.svg?branch=master)](https://coveralls.io/github/MEZk/tg-reminder?branch=master)&nbsp;[![Go Report Card](https://goreportcard.com/badge/github.com/mezk/tg-reminder)](https://goreportcard.com/report/github.com/mezk/tg-reminder)&nbsp;[![Docker Hub](https://img.shields.io/docker/automated/jrottenberg/ffmpeg.svg)](https://hub.docker.com/r/mezk/tg-reminder)

</div>

TG-Reminder is self-hosted reminder bot specifically crafted for Telegram.
Setting it up is straightforward as a Docker container, needing just a Telegram token and user ID for the user to get
started.
Once started, TG-Reminder allows user to create reminders, receiving remind notifications, list, delay, and remove reminders.

## Installation

-   The primary method of installation is via Docker. TG-Reminder is available as a Docker image, making it easy to deploy
    and run as a container. The image is available on Docker Hub
    at [mezk/tg-reminder](https://hub.docker.com/r/mezk/tg-reminder) as well as on GitHub Packages
    at [ghcr.io/mezk/tg-reminder](https://ghcr.io/mezk/tg-reminder).
-   Binary releases are also available on the [releases page](https://github.com/mezk/tg-reminder/releases/latest).
-   TG-Reminder can be installed by cloning the repository and building the binary from source by running `make build`.

## Configuration

All the configuration is done via environment variables:

-   `DB_FILE` – database file path (mandatory)
-   `MIGRATIONS` – migration directory for [goose](https://github.com/pressly/goose), default is `/srv/db/migrations` (optional)
-   `DEBUG` – whether to print debug logs (optional)
-   `TELEGRAM_APITOKEN` – Telegram API token, received from Botfather (mandatory)
-   `TELEGRAM_BOT_API_ENDPOINT` – Telegram API Bot endpoint (optional)
-   `BACKUP_DIR` – directory where to place db backups (optional)
-   `BACKUP_INTERVAL` – how often to make db backups (optional, if BACKUP_DIR is not set)
-   `BACKUP_RETENTION` – retention period for old db backups (optional, if BACKUP_DIR is not set)

## Setting up the telegram bot

To get a token, talk to [BotFather](https://core.telegram.org/bots#6-botfather). All you need is to send `/newbot`
command and choose the name for your bot (it must end in `bot`). That is it, and you got a token which you'll need to
write down into env variable as `TELEGRAM_APITOKEN`.

_Example of such a "talk"_:

```
Andrew:
/newbot

BotFather:
Alright, a new bot. How are we going to call it? Please choose a name for your bot.

Andrew:
example_reminder

BotFather:
Good. Now let's choose a username for your bot. It must end in `bot`. Like this, for example: TetrisBot or tetris_bot.

Andrew:
example_reminder_bot

BotFather:
Done! Congratulations on your new bot. You will find it at t.me/example_reminder_bot. You can now add a description,
about section and profile picture for your bot, see /help for a list of commands. By the way, when you've finished
creating your cool bot, ping our Bot Support if you want a better username for it. Just make sure the bot is fully
operational before you do this.

Use this token to access the HTTP API:
12345678:xy778Iltzsdr45tg
```

## Example of docker-compose.yml

This is an example of a docker-compose.yml file to run the bot. It is using the latest stable version of the bot from docker hub and running as a non-root user with uid:gid 1001:1001 (matching host's uid:gid) to avoid permission issues with mounted volumes. The bot is using UTC timezone.

```yaml
services:
    tg-reminder:
        image: mezk/tg-reminder:latest
        hostname: tg-reminder
        user: "1001:1001" # set uid:gid to host user to avoid permission issues with mounted volumes
        restart: always
        container_name: tg-reminder
        # Allow colorized output
        tty: true
        logging:
            driver: json-file
            options:
                max-size: "10m"
                max-file: "5"
        environment:
            - TZ=UTC
            - TELEGRAM_APITOKEN=${TELEGRAM_APITOKEN}
            - DEBUG=true # if you need debug logs
            - DB_FILE=/srv/db/data/tg-reminder.db # location of database file. We use embedded sqlite.
            - MIGRATIONS=/srv/db/migrations # optional, default directory with migrations is /srv/db/migrations
            - BACKUP_DIR=/srv/db/data/backup # directory where to place db backups
            - BACKUP_RETENTION=48h # retention period for old db backups
            - BACKUP_INTERVAL=12h # how often to make db backups
        volumes:
            - ./var/tg-reminder:/srv/db/data # mount volume with db file
            - ./migrations:/srv/db/migrations # mount volume with db migrations (optional, migrations are copied at docker image build phase into default directory /srv/db/migrations)
```

See [docker-compose.yml](./docker-compose.yml) for more examples.

## Contribution

To start contribution go to issues and choose one.

-   `easy` – to begin contributing;
-   `medium` – if you are already familiar with the codebase;
-   `hard` – if you are good enough to make descisions on how to implement new features and already finished at
    least `medium` issue.

### Makefile

We use `make` to build, test and lint code:

-   `make test` - to run tests and check coverage threshold;
-   `make build` - to build app binary;
-   `make lint` - to run golangci linter;
-   `make docker` - to build Docker image;
-   `make mocks` - to generate mocks.
-   `make bin-deps` - to install required binary dependencies to ./bin.

Don't forget to update `TEST_COVERAGE_THRESHOLD` variable in [Makefile](Makefile) if you increase code coverage.

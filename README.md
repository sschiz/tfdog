# TFDog
[![builds.sr.ht status](https://builds.sr.ht/~mcldresner/tfdog/commits/.build.yml.svg)](https://builds.sr.ht/~mcldresner/tfdog/commits/.build.yml?)

**TFDog** is a Telegram bot that can automatically check for free slots in TestFlight betas.
If the beta has free slots, the bot will notify you about it.

## Installation
Dependencies:
- go
- gcc

```shell
go install git.sr.ht/~mcldresner/tfdog/cmd/tfdog@latest
```

## Usage
```shell
tfdog /path/to/config
```

[Config example](https://git.sr.ht/~mcldresner/tfdog/tree/master/item/examples/config.ini)
## License
AGPLv3, see LICENSE.

Copyright (C) 2021 The tfdog Contributors

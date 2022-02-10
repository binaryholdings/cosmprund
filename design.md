# Design

We want to take in the tendermint state and blocks db store then call prune on them. 

### CMD

flags

- --home = path to db or data directory
- --blocks = amount of blocks to keep in the store
- --versions = amount of versions to keep in the application db
- --backend = db backend
- --db-directory = the dirstory of your db

```sh
cosmprund --home <path_to_db> --blocks 10 --versions 4
```


## Snapshot

- take a snapshot of the highest height locally
- save snapshot locally
- restore local snapshot
- run node
- win

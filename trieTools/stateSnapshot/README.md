# stateSnapshot

Take a snapshot of the current trie state and save it for debugging, testing, or creating checkpoints.

Usage
-----
Example command:

```
./stateSnapshot -log-level *:DEBUG,trie:TRACE -log-save -hex-roothash fd7c9ac71feccdde19d014c3fdacbb46cc36d2c7c9ea583faa7f5c2a1666ec77
```

DB layout
---------
The database directory must follow a sharded layout:

- db/
  - 0  -> sharded trie db
  - 1  -> sharded trie db
  - 2  -> sharded trie db
  - ...
  - n  -> sharded trie db

Behavior
--------
The tool opens all shard databases from 0 up to n, then takes the snapshot in the shard at index n.

Notes
-----
- Provide the desired root hash via `-hex-roothash` to select which state to snapshot.
- Use `-log-save` and `-log-level` to control logging and persist logs if needed.

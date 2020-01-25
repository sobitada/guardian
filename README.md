# Guardian

Checks all `i` milliseconds, whether the Jörumgandr node started with the given node configuration has left the 
bootstrapping phase. If this is the case, then the guardian will undergo three attempts to demote it to a passive node.

**Usage:**
```
Usage: guardian <node-config.yml>
  -i int
    	interval in milliseconds. (default 1000)
```
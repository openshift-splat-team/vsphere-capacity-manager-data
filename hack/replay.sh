
#!/bin/bash

    dlv replay \
    --disable-aslr \
    --backend=rr \
    --headless \
    --listen=:2345 \
    --api-version=2 \
    --accept-multiclient ~/.local/share/rr/latest-trace

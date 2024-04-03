#!/bin/bash
if test -d /run/systemd/system; then
    systemctl daemon-reload
fi

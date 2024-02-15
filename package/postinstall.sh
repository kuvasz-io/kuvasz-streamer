#!/bin/bash
if ! getent group "kuvasz" > /dev/null 2>&1 ; then
    groupadd -r "kuvasz"
fi
if ! getent passwd "kuvasz" > /dev/null 2>&1 ; then
    useradd -r -g kuvasz -d /var/lib/kuvasz -s /sbin/nologin -c "kuvasz user" kuvasz
fi
systemctl daemon-reload

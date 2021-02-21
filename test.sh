#!/bin/bash

for (( ; ; ))
do
    for i in {1..12}
    do

        for (( c=1; c+i<=100; c++ ))
        do
            echo '127.0.0.1 - james [09/May/2018:16:00:39 +0000] "GET /pages HTTP/1.0" 200 123' >> /tmp/access.log

        done

        for  (( c=1; c+i<=90; c++ ))
        do
            echo '127.0.0.1 - james [09/May/2018:16:00:39 +0000] "GET /pages/add HTTP/1.0" 200 344' >> /tmp/access.log
        done

        for  (( c=1; c+i<=80; c++ ))
        do
            echo '127.0.0.1 - james [09/May/2018:16:00:39 +0000] "GET /images HTTP/1.0" 400 33' >> /tmp/access.log
        done

          for  (( c=1; c+i<=70; c++ ))
        do
            echo '127.0.0.1 - james [09/May/2018:16:00:39 +0000] "PUT /document HTTP/1.0" 500 50' >> /tmp/access.log
        done

          for  (( c=1; c+i<=60; c++ ))
        do
            echo '127.0.0.1 - james [09/May/2018:16:00:39 +0000] "PUT /report HTTP/1.0" 303 221' >> /tmp/access.log
        done

             for  (( c=1; c+i<=50; c++ ))
        do
            echo '127.0.0.1 - james [09/May/2018:16:00:39 +0000] "PUT /metrics HTTP/1.0" 404 221' >> /tmp/access.log
        done

        sleep 10

    done

    for i in {1..15}
    do


        for  (( c=1; c+i<=20; c++ ))
        do
            echo '127.0.0.1 - james [09/May/2018:16:00:39 +0000] "PUT /report HTTP/1.0" 303 221' >> /tmp/access.log
        done

        for  (( c=1; c+i<=15; c++ ))
        do
            echo '127.0.0.1 - james [09/May/2018:16:00:39 +0000] "PUT /metrics HTTP/1.0" 404 221' >> /tmp/access.log
        done

        sleep 10

    done

done
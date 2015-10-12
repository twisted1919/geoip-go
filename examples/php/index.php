<?php

// make sure we show any error/warning/notice
error_reporting(-1);
ini_set('display_errors', 1);

function checkIpAddress($ipAddress) {

    if (!filter_var($ipAddress, FILTER_VALIDATE_IP)) {
        throw new Exception("Please provide a valid ip address!");
    }

    $ch = curl_init('http://127.0.0.1:8000/check/' . $ipAddress);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_CONNECTTIMEOUT, 30);
    curl_setopt($ch, CURLOPT_TIMEOUT, 30);

    curl_setopt($ch, CURLOPT_MAXREDIRS, 5);
    curl_setopt($ch, CURLOPT_FOLLOWLOCATION, 1);

    // if password is needed:
    // curl_setopt($ch, CURLOPT_HTTPHEADER, array('Authorization: yourpassword'));

    $body = curl_exec($ch);
    curl_close($ch);

    return (array)json_decode($body, true);
}

print_r(checkIpAddress('123.123.123.123'));

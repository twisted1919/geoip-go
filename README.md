# geoip-go  
A simple GeoIP server written in GO language based on https://github.com/oschwald/geoip2-golang    

### Install  
```
# clone this repository locally (or download it) :  
$ mkdir geoip-go && cd geoip-go && git clone https://github.com/twisted1919/geoip-go.git .  

# install geoip2-golang library:
$ go get github.com/oschwald/geoip2-golang

# fetch latest geoip2 database:  
$ wget http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.mmdb.gz  

# extract the database  
$ gunzip GeoLite2-City.mmdb.gz  

# build the binary:  
$ go build -o geoip-go  

# if needed, edit config.json accordingly
```

### Usage
Start the server with proper flags, use -help to see available options:
```bash
# start server
$ ./geoip-go -database.file="/var/data/GeoLite2-City.mmdb"
```
```php
# client response after sending a GET request to http://localhost:8000/check/123.123.123.123
{
  "status": "success",
  "message": "OK [took 41.622µs]",
  "data": {
    "continent": "Asia",
    "country_name": "China",
    "country_code": "CN",
    "state_name": "Beijing Shi",
    "city_name": "Beijing",
    "postal_code": "",
    "latitude": 39.9289,
    "longitude": 116.3883,
    "timezone": "Asia/Shanghai"
  }
}
```

### Notes  
* command line flags take priority over the ones from configuration file   
* make sure you use -server.password flag to set a password if the server listens on a public interface  


Enjoy.

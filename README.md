[![Build Status](https://travis-ci.org/shadrus/hammer.svg?branch=master)](https://travis-ci.org/shadrus/hammer)

# Hammer
Hammer is a console application for http loading tests.

## Basic usage
1. Load the latest version from Releases.
2. Create config file. It can be in YAML or JSON format.

   ### YAML example
   ```yaml
    url: http://localhost:15500/create
    method: GET
    params:
      url: https://habr.com
    headers:
      User-Agent: hammer
      Content-Type: application/json
    body: '{"test": 3}'
    timeout: 2
    steps: 3
    maxRPS: 5
    growsCoef: 0.8
    basicauth:
      testuser: testpassword
    ```
     ### JSON example
     ``` javascript
     {
         "url": "http://localhost:15500/create",
         "method": "GET",
         "params":{"url: https://habr.com"},
         "headers": {"User-Agent": "hammer"},
         "timeout": 2,
         "steps": 3,
         "maxRPS": 5,
         "growsCoef": 0.8
    }
    ```

    #### Possible parameters
    a. **url** - what to test
    
    b. **method** - request method
    
    c. **params** - request parameters
    
    d. **headers** - request headers
    
    e. **timeout** - request timeout to prevent too long requests

    f. **steps** - how many iterations will be performed
    
    g. **basicauth** - username and password for the sites where must be BasicAuth    
    
    h. **body** - request body    
  
    i. **maxRPS** - loading presser starts from 1 request per step and to maxRPS
    
    j **growsCoef** - if we have durability = 100 than with growsCoef = 0.8 we'll have maxRPS after 80 seconds

1. Run program

## Additional console parameters
**-conf=configfilename.yaml(.json)** required parameter with task configuration

**-d** DEBUG mode will show debugging logs

**-out=reportname.csv** to generate report based on created requests.

**--help** show help






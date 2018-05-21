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
    durability: 3
    maxRPS: 5
    growsCoef: 0.8
    basicauth:
    testuser: testpassword
    ```
     ### JSON example
     ``` javascript
     {"url": "http://localhost:15500/create",
      "method": "GET",
      "params":{"url: https://habr.com"},
      "headers": {"User-Agent": "hammer"},
      "timeout": 2,
      "durability": 3,
      "maxRPS": 5,
      "growsCoef": 0.8}
    ```

3. Run program with param -conf=configfilename.yaml(.json)

You can also use param -out=reportname.csv to generate report based on created requests.






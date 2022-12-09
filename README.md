# proxy-checker

## Steps to run:
- Clone this repository using:
    `git clone https://github.com/44za12/proxy-checker.git`
- Run:
    - `cd proxy-checker`
    - `go build -o proxychecker`
    - `./proxychecker all.txt good.txt`

Running the above commands will generate two text files: `all.txt` and `good.txt` `all.txt` has all the proxies that were scraped and `good.txt` has the filtered _good_ proxies.
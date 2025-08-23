### About
This project attempts to implement Windows service similiar to unix logrotate as windows doesn't really have a valid equivalent.

### Features
- Delete or rotate logs (or any files really)
- Option to delete/rotate conditionally based on file age, size, or time interval
- Compression for rotated files in gzip/zip
- Support for pre/post custom scripts

### Configuration
See configs/wingologrotate.yaml for example config.

### Usage
- place exe in desired directory
- create configs/wingologrotate.yaml in same location
- run wingologrotate.exe install as administrator
- start the windows service

### Compatibility
Tested on Windows 10 and Windows Server 2019. It should run on any modern windows distribution.
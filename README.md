### About
This project attempts to implement service similiar to unix logrotate as windows doesn't really have a valid equivalent.

### Features
- Delete or rotate logs (or any files really)
- Option to delete/rotate conditionally based on file age or size
- Compression for rotated files in gzip/zip
- TODO delete/rotate on time interval condition
- TODO pre/post custom script

### Configuration
See configs/wingologrotate.yaml for example config.

### Usage
- place exe in desired directory
- create configs/wingologrotate.yaml in same location
- run wingologrotate.exe install as administrator
- start the windows service

### Compatibility
Tested on Windows 10 and Windows Server 2019. It should run on any modern windows distribution.
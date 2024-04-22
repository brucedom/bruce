==== BRUCE ====

Basic runtime for uniform compute environments

Bruce was initially intended to just operate as a more advanced exec handler for serf.  It has somewhat evolved at this point far beyond that in order to become a more stable OS configuration and installation utility.  More stable and capable as it does not require pre-existing libraries on the base OS like ansible, or agents that must be configured and associated with a chef server etc.  One of the key characteristics is the ability to load templates directly through multiple loaders.  This enables not only the ability to quickly setup a fleet by hostname as example but also to effectively bootstrap an instance on EC2 in a secure way by limiting that particular instance profile to an s3 prefix from which bruce will load the installer config.

## TLDR How do I run bruce.

### A one liner to download latest bruce:
```
wget -qO- $(curl -s https://api.github.com/repos/brucedom/bruce/releases/latest | grep "linux_amd64"|grep https | cut -d : -f 2,3 | tr -d \" | awk '{$1=$1};1') |tar -xvz
```

Once you've downloaded your respective OS package from: https://github.com/brucedom/bruce/releases/latest

Extract it and run bruce with a config file, for an example config file see: https://github.com/brucedom/bruce/blob/main/config.example.yml

```
./bruce --config=install.yml
```

Or in the event you want to load it from an instance that should load an internal s3 hosted install file:
```
./bruce --config s3://somebucket/install.yml
```

Or if you prefer to have an internal service that hosts all your files:
```
./bruce --config https://some.hostname/$(hostname -f).yml
```

Currently bruce supports several operators within the config file that provide the functionality:
* Native commands with built in os limiters (so you can limit which OS's will run what commands) - see nginx example
* Services which will enable services and will auto restart services based on templates that trigger restarts during a run (can be used with serf to auto update)
* Packages which will install OS packages on the host system to configure the system for use
* Ownership to enable chowning one file or recursive directories of files
* Signals in order to send SIGINT / SIGHUP to running processes instead of restarting the entire process
* Templates which support injection of variables via locally run commands as input value and provided template values
* Several more operators to come.

Principles for building bruce & why not ansible?:
- NO additional OS dependencies, should be able to use it on scratch if I want...
- Single binary (aka go binary)
- Multi platform (aka linux / mac / [basic windows support already])
- Must do package installs (at least yum & apt for now)
- Must configure templates (concurrently if possible)
- Must be way faster than ansible IE: configure entire system before checks pass on an amazon t2.micro for general installs like nginx

===== Extended Functionality =====
- Template variables are fully injected into the template system and the ability to set envars on the fly exists, benefit here is that those envars only exist during the session run making it beneficial even for configuring connections with passwords.
- Download and extract tarball to specified directory, stripping of initial directories inside also exists making it simple to handle package releases of applications.
- Execute or exclude use of commands based on conditions.
- Basic windows functionality but requires additional sourcing from the community to make it a fully baked solution.
- Run as a server, enable the ability to trigger runs remotely through a basic GET request reducing the need for login credentials.
- Restart services only on change detection.

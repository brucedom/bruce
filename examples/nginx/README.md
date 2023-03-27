# NGINX Example

We'll use AWS Ec2 to spin up a t2 micro and use the ec2-user data to do initial download of bruce.
Bruce will use the install.yml file within this directory to install nginx and configure it with a default vhost
which redirects http -> https & also updates the standard html under /var/www/html/index.hml with some reconfigured variables.

NOTE: templates within this example should match directory wise with what will be created on the server to make it easier to inspect.

I used some basic terraform to create a generic VPC with a single subnet (look at TF on how to do it or make one yourself):
This example should work fine on an intel nuc etc.  To understand how that is done take a look at the ec2-userdata.txt file in this directory.

# Step 1

With the repository cloned you can make use of the terraform code within this directory to spin up your own ec2 environment or just copy the ec2-userdata.sh contents and put it as part of an existing ec2 instance that you will spin up.
The terraform code that is here as example will provision a vpc within a single region with a single public facing instance and include the userdata file to be run on instance initialization.
The instance type is a t2.micro instance.  Alternatively you can run the user-data script directly on an intel nuc loaded with fedora / ubuntu as example.


# Step 2
Validate what is happening on the ec2 instance by logging in (terraform code should output connect string)
### Note: For examples I will use Key Name: Nitecon provision your own ssh key during tf instantiation or / use existing keys as appropriate.

Once logged into the system take a look at your cloudinit log details by running for example: 
```
cat /var/log/cloud-init-output.log
```


# Step 3
After terraform completed and the instance stabilizes bruce should have installed everything on the system and you should be able to hit your public innstance via a browser.

- Enjoy
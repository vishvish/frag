# FRAG

## a FreeAgent CLI utility

I am bored of faffing around with the FreeAgent GUI, it's slow and repetitive.

So I'm making this CLI tool to help me deal with some of these repetitive tasks, like expenses and transaction categorization.

I'm only writing it for Mac, but it'll probably work on other platforms.


## Usage

    frag login      # fires up browser to authorize app and obtain token
    frag get        # sample command to fetch project list


Tokens are stored in the Mac OS Keychain. If a token is not found, frag exits and asks you to login instead.


## Tings

It's written in Golang using the spf13/cobra cli framework. This is a bit weird and probably over-engineered but I've gotten it to work. I was curious as to how it worked mainly. Maybe I'll switch to mitchellh/cli which is used for Vagrant, Consul and all the Hashicorp stuff. Mitchell is a nice guy, too.

Designed to work on Mac OS. Uses the keyring designed for `aws-vault` so should be fairly solid if you want to turn it into something cross-platform.

You will need two Environment variables for this to work. FREEAGENT_CLIENT_ID and FREEAGENT_CLIENT_SECRET.

Go get yourself a new app from https://dev.freeagent.com/

No Guarantee, No Tests, and GPL3.
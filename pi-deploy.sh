# This script deploys the current version the the pi.
# Therefore it compiles the bot and then logs into the pi via ssh and swaps the old for the new bot.
# This script only works for the Raspberry Pi 1 B+. To adapt it to other targets you need to modify the the build
# command.

echo "--- Brobot delpoyment script ---"
echo "Target: Raspberry Pi 1 B+"
echo ""
echo "The script requires sshpass to be installed on your system"
echo "https://gist.github.com/arunoda/7790979"
echo ""

Remote="pi@192.168.8.164"

echo "Password for $Remote: "
read -s -r Password

echo "Building the bot ..."
# TODO: Change this to match your rapsbery pi
# GOARM=6 if Raspberry Pi 1
# GOARM=7 if Raspberry Pi 2
# GOARM=8 if Raspberry Pi 3 or newer
env GOOS=linux GOARCH=arm GOARM=6 go build

echo "Stoping the old bot ..."
sshpass -p "$Password" ssh $Remote "sudo systemctl stop brobot;"

echo "Uploading the new bot ..."
sshpass -p "$Password" scp ./brobot $Remote:/home/pi/brobot/

echo "Starting the new bot ..."
sshpass -p "$Password" ssh $Remote "sudo systemctl daemon-reload; sudo systemctl start brobot;"

# Delete the binary for the raspberry because it is confusing to have a binary that doesn't work on this system
echo "Cleaning up ..."
rm ./brobot

echo ""
echo "Done :)"

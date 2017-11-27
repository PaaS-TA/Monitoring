#!/usr/bin/env bats

load helpers

function setup() {
	teardown_busybox
	setup_busybox
}

function teardown() {
	teardown_busybox
}

@test "runc run [tty ptsname]" {
	# Replace sh script with readlink.
    sed -i 's|"sh"|"sh", "-c", "for file in /proc/self/fd/[012]; do readlink $file; done"|' config.json

	# run busybox
	runc run test_busybox
	[ "$status" -eq 0 ]
	[[ ${lines[0]} =~ /dev/pts/+ ]]
	[[ ${lines[1]} =~ /dev/pts/+ ]]
	[[ ${lines[2]} =~ /dev/pts/+ ]]
}

@test "runc run [tty owner]" {
	# tty chmod is not doable in rootless containers.
	# TODO: this can be made as a change to the gid test.
	requires root

	# Replace sh script with stat.
	sed -i 's/"sh"/"sh", "-c", "stat -c %u:%g $(tty) | tr : \\\\\\\\n"/' config.json

	# run busybox
	runc run test_busybox
	[ "$status" -eq 0 ]
	[[ ${lines[0]} =~ 0 ]]
	# This is set by the default config.json (it corresponds to the standard tty group).
	[[ ${lines[1]} =~ 5 ]]
}

@test "runc run [tty owner] ({u,g}id != 0)" {
	# tty chmod is not doable in rootless containers.
	requires root

	# replace "uid": 0 with "uid": 1000
	# and do a similar thing for gid.
	sed -i 's;"uid": 0;"uid": 1000;g' config.json
	sed -i 's;"gid": 0;"gid": 100;g' config.json

	# Replace sh script with stat.
	sed -i 's/"sh"/"sh", "-c", "stat -c %u:%g $(tty) | tr : \\\\\\\\n"/' config.json

	# run busybox
	runc run test_busybox
	[ "$status" -eq 0 ]
	[[ ${lines[0]} =~ 1000 ]]
	# This is set by the default config.json (it corresponds to the standard tty group).
	[[ ${lines[1]} =~ 5 ]]
}

@test "runc exec [tty ptsname]" {
	# run busybox detached
	runc run -d --console-socket $CONSOLE_SOCKET test_busybox
	[ "$status" -eq 0 ]

	# make sure we're running
	testcontainer test_busybox running

	# run the exec
    runc exec test_busybox sh -c 'for file in /proc/self/fd/[012]; do readlink $file; done'
	[ "$status" -eq 0 ]
	[[ ${lines[0]} =~ /dev/pts/+ ]]
	[[ ${lines[1]} =~ /dev/pts/+ ]]
	[[ ${lines[2]} =~ /dev/pts/+ ]]
}

@test "runc exec [tty owner]" {
	# tty chmod is not doable in rootless containers.
	# TODO: this can be made as a change to the gid test.
	requires root

	# run busybox detached
	runc run -d --console-socket $CONSOLE_SOCKET test_busybox
	[ "$status" -eq 0 ]

	# make sure we're running
	testcontainer test_busybox running

	# run the exec
    runc exec test_busybox sh -c 'stat -c %u:%g $(tty) | tr : \\n'
	[ "$status" -eq 0 ]
	[[ ${lines[0]} =~ 0 ]]
	[[ ${lines[1]} =~ 5 ]]
}

@test "runc exec [tty owner] ({u,g}id != 0)" {
	# tty chmod is not doable in rootless containers.
	requires root

	# replace "uid": 0 with "uid": 1000
	# and do a similar thing for gid.
	sed -i 's;"uid": 0;"uid": 1000;g' config.json
	sed -i 's;"gid": 0;"gid": 100;g' config.json

	# run busybox detached
	runc run -d --console-socket $CONSOLE_SOCKET test_busybox
	[ "$status" -eq 0 ]

	# make sure we're running
	testcontainer test_busybox running

	# run the exec
    runc exec test_busybox sh -c 'stat -c %u:%g $(tty) | tr : \\n'
	[ "$status" -eq 0 ]
	[[ ${lines[0]} =~ 1000 ]]
	[[ ${lines[1]} =~ 5 ]]
}

@test "runc exec [tty consolesize]" {
	# allow writing to filesystem
	sed -i 's/"readonly": true/"readonly": false/' config.json

	# run busybox detached
	runc run -d --console-socket $CONSOLE_SOCKET test_busybox
	[ "$status" -eq 0 ]

	# check state
	wait_for_container 15 1 test_busybox

	# make sure we're running
	testcontainer test_busybox running

# write a process.json that prints tty info to a file
cat <<EOF > $BATS_TMPDIR/write-tty-info.json
{
	"terminal": true,
	"consoleSize": {
		"height": 10,
		"width": 110
	},
	"args": [
		"/bin/sh",
		"-c",
		"/bin/stty -a > $BATS_TMPDIR/tty-info"
	],
        "cwd": "/"
}
EOF

	# run the exec
	runc exec --pid-file pid.txt -d --console-socket $CONSOLE_SOCKET -p $BATS_TMPDIR/write-tty-info.json test_busybox
	[ "$status" -eq 0 ]

	# check the pid was generated
	[ -e pid.txt ]

	#wait user process to finish
	timeout 1 tail --pid=$(head -n 1 pid.txt) -f /dev/null

# write a process.json that echoes the tty file info
cat <<EOF >> $BATS_TMPDIR/read-tty-info.json
{
	"args": [
	    "/bin/cat",
	    "$BATS_TMPDIR/tty-info"
	],
        "cwd": "/"
}
EOF

	# run the exec
	runc exec -p $BATS_TMPDIR/read-tty-info.json test_busybox
	[ "$status" -eq 0 ]

	# test tty width and height against original process.json
	[[ ${lines[0]} =~ "rows 10; columns 110" ]]

	# clean process.jsons
	rm $BATS_TMPDIR/write-tty-info.json
	rm $BATS_TMPDIR/read-tty-info.json
}

@test "runc exec [tty consolesize] (width && height > max)" {
	# run busybox detached
	runc run -d --console-socket $CONSOLE_SOCKET test_busybox
	[ "$status" -eq 0 ]

	# check state
	wait_for_container 15 1 test_busybox

	# make sure we're running
	testcontainer test_busybox running

# write a process.json that prints tty info to a file
cat <<EOF > $BATS_TMPDIR/process.json
{
	"terminal": true,
	"consoleSize": {
		"height": 100000,
		"width": 500000
	},
	"args": [
		"/bin/true"
	],
        "cwd": "/"
}
EOF

	# run the exec
	runc exec --pid-file pid.txt -d --console-socket $CONSOLE_SOCKET -p $BATS_TMPDIR/process.json test_busybox
	[ "$status" -ne 0 ]
	[[ ${lines[2]} =~ "console width (500000) or height (100000) cannot be larger than 65535" ]]
}

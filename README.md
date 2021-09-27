# Deep directories in containers

The depth of directory trees is not bounded in some container platforms. This allows a maliicious container processes to create directory trees that, depending on platform and system resources, can exhaust available memory when traversing, making removal difficult without manual intervention.

## Demo

Create a pod that creates a very deep series of nested directories:

```
$ kubectl create pod.yml
```

Monitor logs for progress:

```
$ kubectl logs -f deep
0
128
256
384
512
...
1056128
```
Eventually the pod will be killed, and recreated. At the same time one can monitor resource usage with with top/htop/slabtop etc. Depending on alloted resources, one will observe a spike in either CPU usage, memory usage or both.

Log entries like below will be found:

```
$ journalctl -fu kubelet
Sep 27 06:41:39 minikube kubelet[4049]: E0927 06:41:39.079981    4049 fsHandler.go:114] failed to collect filesystem stats - rootDiskErr: unable to count inodes for part of dir /var/lib/docker/overlay2/de1eb00a41a325ce94b7b2f7f034d86959edb022b5c652f07b521ec5e722efaf/diff: lstat /var/lib/docker/overlay2/de1eb00a41a325ce94b7b2f7f034d86959edb022b5c652f07b521ec5e722efaf/diff/tmp/deep/x/x/x/x/x/x/x/x/x/x/x/x/x...
```

The pod will never successfully be removed, it's file tree will remain on the node. The directory tree can be removed manually with `rm -rf`, if sufficient memory/swap is available (otherwise it will be killed by OOM killer) though can take several minutes to complete. If the pod is not manually deleted, then the node will eventually become unstable as repeated failed attempts to cleanup the pod continually trigger OOM killer:

```
$ dmesg | grep -i killed
...
[ 2030.962075] Killed process 40422 (containerd) total-vm:1121000kB, anon-rss:22896kB, file-rss:0kB, shmem-rss:28268kB
[ 2031.299167] Killed process 41907 ((agetty)) total-vm:176776kB, anon-rss:2976kB, file-rss:0kB, shmem-rss:44kB
[ 2031.331869] Killed process 41908 (kubelet) total-vm:118092kB, anon-rss:132kB, file-rss:0kB, shmem-rss:1152kB
[ 2031.339795] Killed process 40414 (dockerd) total-vm:6866100kB, anon-rss:2112192kB, file-rss:0kB, shmem-rss:45740kB
...
```

On minikube, the cluster will become completely unresponsive.


## Container cleanup

Most container runtimes are written in Go. The os.RemoveAll() function from the Go standard library has some additional memory overhead (when compared to coreutils' rm) when traversing very deep directories. 

https://github.com/golang/go/issues/47390

Thus runtimes like CRI-O, which use this function to clean up containers, can struggle with removal of containers that have created very deep directories, repeatedly triggering OOM killer, potentially putting the system into an unstable state.

## CRI-O (and podman, buildah etc)

This has been reported to CRI-O here:

https://github.com/cri-o/cri-o/issues/5126

Support for inode quotas a container's read/write layer (in xfs, the default filesystem on RHEL) was recently added to the containers/storage library which is used by CRI-O, podman, buildah and skopeo:

https://github.com/containers/storage/pull/970

## Kubernetes

Similarly impacted and reported to their security team. The Kubernetes SRC are comfortable with a fix for this issue to be investigated and worked on in the open.

Kubernetes does not yet have controls to limit the number of inodes used by a container. This would appear the most viable approach to prevent hostile containers from creating directory trees of this size.

## Docker / Containerd

Similarly impacted and privately reported to their security team.

## Acknowledgments

This behaviour was observed after investigating the impact to container platforms of prior kernel CVE, CVE-2021-33909, discovered by the Qualys Research Team.

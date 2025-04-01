# Customized backup/restore controller
This is a kubernetes controller that is going to backup the volumes, using kubernetes snapshot APIs, that are provided by the users as input. Once the volume is snapshotted users can also choose to restore that snapshot to get their data back.

## Description
This controller allow users to select volume to take the backup and restore it whenever needed. In the backend, k8s snapshot APIs are being used to create storage resources. 

## Design 
Both the backup and restore controller follow controller pattern. The backup controller watches for userbackup CR create/update/delete events. Backup controller creates/deletes VolumeSnapshot object accordingly. The restore controller watches for userrestore CR create/update/delete events. Restore controller creates/deletes PVC object accordingly.

There are two more controller for reflecting the status of underline resources. Backup_status controller watches on VolumeSnapshot CR and syncs userBackup resource status with the VolumeSnapshot CR. Restore_status controller watches on PVC CR and syncs userRestore resource status with the PVC CR. 

Above controllers work together to give seamless experience. Once userBackup resource is created, backup controller sets initial conditions and creates underline Volume Snapshot CR. K8s VolumeSnapshot controller reconciles volumesnapshot CR. UserBackupStatus controller watching on VolumeSnapshot syncs the status of userbackup resource with volumesnapshot status. 

Similarly, userRestore resource is created, restore controller sets initial conditions and creates underline PVC CR. K8s storage controller reconciles PVC. UserRestoreStatus controller watching on PVC syncs the status of userrestore resource with PVC status.

## Usage

### Prerequisites
- go version v1.23.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### Set up dev environment. 
1. Create kind cluster. 
    ```sh
    kind create cluster --name my-kind-cluster --image kindest/node:v1.32.3
    ```
2. Install snapshot controller and CRDs. 
    ```sh
    git clone https://github.com/kubernetes-csi/external-snapshotter.git
    cd external-snapshotter
    Add --feature-gates=CSIVolumeGroupSnapshot=true in the deploy/kubernetes/snapshot-controller/setup-snapshot-controller.yaml
    kubectl kustomize https://github.com/kubernetes-csi/external-snapshotter/client/config/crd | kubectl create -f -
    kubectl -n kube-system kustomize deploy/kubernetes/snapshot-controller | kubectl create -f -
    ```
3. Install CSI Driver.
    ```sh
    git clone https://github.com/kubernetes-csi/csi-driver-host-path.git
    cd csi-driver-host-path
    ./deploy/kubernetes-latest/deploy.sh
    ```

### Steps for installing customized backup/restore controller:- 
1. Clone repo locally 
    ```sh
    git clone https://github.com/alka-infracloud/storage-customized-controller.git
    ```

2. Install CRDs
    ```sh
    kubectl apply  -f config/crd/bases
    ```

3. Deploy to the cluster.
    ```sh
    make install
    make deploy IMG=calka/customized_backup_restore_ctrl:latest
    ```

4. User can snapshot a PVC by following the steps mentioned below. 
You can apply the samples (examples) from the config/sample:
    ```sh
   kubectl apply -f config/samples/userbackup.yaml
   kubectl apply -f config/samples/userrestore.yaml
   ```

5. User can delete the snapshot by following the steps mentioned below. 
    ```sh
    kubectl apply -k config/samples/
    ```

6. Delete the APIs(CRDs) from the cluster.

    ```sh
    make uninstall
    ```

7. Undeploy the controller from the cluster.

    ```sh
    make undeploy
    ```

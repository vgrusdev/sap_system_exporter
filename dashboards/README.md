# Grafana dashboards

We provide a dashboards for Grafana, leveraging the exporter.


## SAP NetWeaver

This dashboard shows the details of a single landscape.

![SAP NetWeaver](screenshot.png)


## Installation

### RPM 

On openSUSE and SUSE Linux Enterprise distributions, you can install the package via zypper in your Grafana host:
```
zypper in grafana-sap-netweaver-dashboards
systemctl restart grafana-server
```

For the latest development version, please refer to the [development upstream project in OBS](https://build.opensuse.org/project/show/network:ha-clustering:sap-deployments:devel), which is automatically updated everytime we merge changes in this repository. 

### Manual

Copy the [provider configuration file](https://build.opensuse.org/package/view_file/network:ha-clustering:sap-deployments:devel/grafana-sap-providers/provider-sles4sap.yaml?expand=1) in `/etc/grafana/provisioning/dashboards` and then the JSON files inside `/var/lib/grafana/dashboards/sles4sap`.

Once done, restart the Grafana server.


## Development notes

- Please make sure the `version` field in the JSON is incremented just once per PR.
- Unlike the exporter, OBS Submit Requests are not automated for the dashboard package.  
  Once PRs are merged, you will have to manually perform a Submit Request against `openSUSE:Factory`, after updating the `version` field in the `_service` file and adding an entry to the `grafana-sap-netweaver-dashboards.changes` file.    

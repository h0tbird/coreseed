---
title: Services and ports
---

# Services and ports

<br>

<div class="panel panel-default">
  <div class="panel-heading"><span class="glyphicon glyphicon-cog" aria-hidden="true"></span> Listen ports used by <b>all</b> roles</div>
  <table class="table">
   <thead><tr> <th>Service</th> <th>Ports</th> <th>Protocol</th> <th>Access</th> <th>Roles</th> </tr></thead>
   <tbody>
   <tr> <td>rexray</td> <td>7979</td> <td>tcp</td> <td>internal</td> <td>all</td> </tr>
   <tr> <td>etcd2</td> <td>2379</td> <td>tcp</td> <td>internal</td> <td>all</td> </tr>
   <tr> <td>node-exporter</td> <td>9101</td> <td>tcp</td> <td>internal</td> <td>all</td> </tr>
   <tr> <td>cadvisor</td> <td>4194</td> <td>tcp</td> <td>internal</td> <td>all</td> </tr>
   <tr> <td>systemd</td> <td>22</td> <td>tcp</td> <td>internal</td> <td>all</td> </tr>
   </tbody>
  </table>
</div>

<div class="panel panel-default">
  <div class="panel-heading"><span class="glyphicon glyphicon-cog" aria-hidden="true"></span> Listen ports used by <b>quorum</b> role</div>
  <table class="table">
   <thead><tr> <th>Service</th> <th>Ports</th> <th>Protocol</th> <th>Access</th> <th>Roles</th> </tr></thead>
   <tbody>
   <tr> <td>etcd2</td> <td>2380</td> <td>tcp</td> <td>internal</td> <td>quorum</td> </tr>
   <tr> <td>zookeeper</td> <td>2888,3888,2181</td> <td>tcp</td> <td>internal</td> <td>quorum</td> </tr>
   <tr> <td>zookeeper-exporter</td> <td>9103</td> <td>tcp</td> <td>internal</td> <td>quorum</td> </tr>
   </tbody>
  </table>
</div>

<div class="panel panel-default">
  <div class="panel-heading"><span class="glyphicon glyphicon-cog" aria-hidden="true"></span> Listen ports used by <b>master</b> role</div>
  <table class="table">
   <thead><tr> <th>Service</th> <th>Ports</th> <th>Protocol</th> <th>Access</th> <th>Roles</th> </tr></thead>
   <tbody>
   <tr> <td>mesos-exporter</td> <td>9104</td> <td>tcp</td> <td>internal</td> <td>master</td> </tr>
   <tr> <td>marathon</td> <td>8080,9292</td> <td>tcp</td> <td>internal</td> <td>master</td> </tr>
   <tr> <td>mesos-dns</td> <td>53</td> <td>tcp</td> <td>internal</td> <td>master</td> </tr>
   <tr> <td>mesos-master</td> <td>5050</td> <td>tcp</td> <td>internal</td> <td>master</td> </tr>
   <tr> <td>prometheus</td> <td>9191</td> <td>tcp</td> <td>internal</td> <td>master</td> </tr>
   <tr> <td>alertmanager</td> <td>9093</td> <td>tcp</td> <td>internal</td> <td>master</td> </tr>
   </tbody>
  </table>
</div>

<div class="panel panel-default">
  <div class="panel-heading"><span class="glyphicon glyphicon-cog" aria-hidden="true"></span> Listen ports used by <b>worker</b> role</div>
  <table class="table">
   <thead><tr> <th>Service</th> <th>Ports</th> <th>Protocol</th> <th>Access</th> <th>Roles</th> </tr></thead>
   <tbody>
   <tr> <td>marathon-lb</td> <td>9090,9091</td> <td>tcp</td> <td>internal</td> <td>worker</td> </tr>
   <tr> <td>marathon-lb</td> <td>80,443</td> <td>tcp</td> <td><font color="green">external</font></td> <td>worker</td> </tr>
   <tr> <td>go-dnsmasq</td> <td>53</td> <td>tcp</td> <td>internal</td> <td>worker</td> </tr>
   <tr> <td>mesos-slave</td> <td>5051</td> <td>tcp</td> <td>internal</td> <td>worker</td> </tr>
   <tr> <td>haproxy-exporter</td> <td>9102</td> <td>tcp</td> <td>internal</td> <td>worker</td> </tr>
   <tr> <td>mesos-exporter</td> <td>9105</td> <td>tcp</td> <td>internal</td> <td>worker</td> </tr>
   </tbody>
  </table>
</div>

<div class="panel panel-default">
  <div class="panel-heading"><span class="glyphicon glyphicon-cog" aria-hidden="true"></span> Listen ports used by <b>border</b> role</div>
  <table class="table">
   <thead><tr> <th>Service</th> <th>Ports</th> <th>Protocol</th> <th>Access</th> <th>Roles</th> </tr></thead>
   <tbody>
   <tr> <td>mongodb</td> <td>27017</td> <td>tcp</td> <td>internal</td> <td>border</td> </tr>
   <tr> <td>pritunl</td> <td>9756</td> <td>tcp</td> <td>internal</td> <td>border</td> </tr>
   <tr> <td>pritunl</td> <td>18443</td> <td>udp</td> <td><font color="green">external</font></td> <td>border</td> </tr>
   <tr> <td>pritunl</td> <td>80,443</td> <td>tcp</td> <td><font color="green">external</font></td> <td>border</td> </tr>
   </tbody>
  </table>
</div>

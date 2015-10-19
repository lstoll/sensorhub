# itron-sdr

This reads data from itron meters, and reports it to librato. Uses [rtlamr](https://github.com/bemasher/rtlamr) for the heavy lifting.

Installation: make sure you have rtlamr available. The deb provided will pull this in

Running: Start rtl_tcp in the background. Then run this. Use `-help` for the options. All fields are required except for gas price. If you do specify a gas price, that will get reported too.

Following is what I use in librato:

![screenshot of graphs](http://cdn.lstoll.net/screen/Tremont_HVAC__Librato_2015-10-19_10-48-25.png)

* cf/hr consumption: `derive(s("sensor.utility.gas.cf", "tremont", { period:"3600", function:"sum" }))`
* $/hr spend: `derive(s("sensor.utility.gas.spend", "tremont", { period:"3600", function:"sum" }))`
* weekly rolling spend: `timeshift("0,1,2,3w", integrate(sum(s("sensor.utility.gas.spend", "tremont"))))`

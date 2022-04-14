$(setupTravelMap());

function setupTravelMap() {
  var map = L.map('map', {
    fullscreenControl: true
  });

  L.tileLayer('https://api.mapbox.com/styles/v1/{id}/tiles/{z}/{x}/{y}?access_token={accessToken}', {
    attribution: 'Map data &copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors, Imagery © <a href="https://www.mapbox.com/">Mapbox</a>',
    maxZoom: 18,
    id: 'mapbox/outdoors-v11',
    tileSize: 512,
    zoomOffset: -1,
    accessToken: 'pk.eyJ1Ijoid2lsbGZsZWV0dyIsImEiOiJjbDF5Y3dndWkwYmgwM2NwbTVjbnNidjI4In0.G3Em4atmvBTTDWyg4_UhRg'
  }).addTo(map);

  var latlngs = [
    [45.51, -122.68],
    [37.77, -122.43],
    [34.04, -118.2]
  ];

  var polyline = L.polyline(latlngs, {color: 'red'}).addTo(map);

  // zoom the map to the polyline
  map.fitBounds(polyline.getBounds());
}
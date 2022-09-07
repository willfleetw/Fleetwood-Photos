const map = L.map('map');
L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
  maxZoom: 19,
  attribution: '© OpenStreetMap'
}).addTo(map);
map.on('click', onMapClick);
let marker = L.marker([50.5, 30.5]).addTo(map);
marker.setOpacity(0);

let iconCodes = [];
(() => {
  iconCodes[200] = "thunderstorm";
  iconCodes[201] = "thunderstorm";
  iconCodes[202] = "thunderstorm";
  iconCodes[210] = "lightning";
  iconCodes[211] = "lightning";
  iconCodes[212] = "lightning";
  iconCodes[221] = "lightning";
  iconCodes[230] = "thunderstorm";
  iconCodes[231] = "thunderstorm";
  iconCodes[232] = "thunderstorm";
  iconCodes[300] = "sprinkle";
  iconCodes[301] = "sprinkle";
  iconCodes[302] = "rain";
  iconCodes[310] = "rain-mix";
  iconCodes[311] = "rain";
  iconCodes[312] = "rain";
  iconCodes[313] = "showers";
  iconCodes[314] = "rain";
  iconCodes[321] = "sprinkle";
  iconCodes[500] = "sprinkle";
  iconCodes[501] = "rain";
  iconCodes[502] = "rain";
  iconCodes[503] = "rain";
  iconCodes[504] = "rain";
  iconCodes[511] = "rain-mix";
  iconCodes[520] = "showers";
  iconCodes[521] = "showers";
  iconCodes[522] = "showers";
  iconCodes[531] = "storm-showers";
  iconCodes[600] = "snow";
  iconCodes[601] = "snow";
  iconCodes[602] = "sleet";
  iconCodes[611] = "rain-mix";
  iconCodes[612] = "rain-mix";
  iconCodes[615] = "rain-mix";
  iconCodes[616] = "rain-mix";
  iconCodes[620] = "rain-mix";
  iconCodes[621] = "snow";
  iconCodes[622] = "snow";
  iconCodes[701] = "showers";
  iconCodes[711] = "smoke";
  iconCodes[721] = "day-haze";
  iconCodes[731] = "dust";
  iconCodes[741] = "fog";
  iconCodes[761] = "dust";
  iconCodes[762] = "dust";
  iconCodes[771] = "cloudy-gusts";
  iconCodes[781] = "tornado";
  iconCodes[800] = "day-sunny";
  iconCodes[801] = "cloudy-gusts";
  iconCodes[802] = "cloudy-gusts";
  iconCodes[803] = "cloudy-gusts";
  iconCodes[804] = "cloudy";
  iconCodes[900] = "tornado";
  iconCodes[901] = "storm-showers";
  iconCodes[902] = "hurricane";
  iconCodes[903] = "snowflake-cold";
  iconCodes[904] = "hot";
  iconCodes[905] = "windy";
  iconCodes[906] = "hail";
  iconCodes[957] = "strong-wind";
})();
geolocate();

async function updateLocation(latitude, longitude) {
  let zoomLevel = 13;
  if (typeof map.getZoom() !== "undefined") {
    zoomLevel = map.getZoom();
  }
  map.setView([latitude, longitude], zoomLevel);

  marker.setLatLng([latitude, longitude]);
  marker.setOpacity(1);

  let url = `https://api.openweathermap.org/data/2.5/weather?lat=${latitude}&lon=${longitude}&appid=774e68a2ef152559f5a0f30f246938cd`;
  respJson = await (await fetch(url)).json();

  let main = respJson.main;
  let weather = respJson.weather[0];

  $('#weather-icon').attr("class", "wi wi-" + iconCodes[weather.id]);

  const times = SunCalc.getTimes(new Date(), latitude, longitude);
  $('#sunrise-time').text(times.sunrise.toLocaleTimeString("en-US", {timeStyle: "short"}));
  $('#sunset-time').text(times.sunset.toLocaleTimeString("en-US", {timeStyle: "short"}));

  let goldenHourMorningStart = new Date(times.goldenHourEnd);
  goldenHourMorningStart.setHours(goldenHourMorningStart.getHours() - 1);
  $('#goldenhour-morning').text(`${goldenHourMorningStart.toLocaleTimeString("en-US", {timeStyle: "short"})} - ${times.goldenHourEnd.toLocaleTimeString("en-US", {timeStyle: "short"})}`);

  let goldenHourEveningEnd = new Date(times.goldenHour);
  goldenHourEveningEnd.setHours(goldenHourEveningEnd.getHours() + 1);
  $('#goldenhour-evening').text(`${times.goldenHour.toLocaleTimeString("en-US", {timeStyle: "short"})} - ${goldenHourEveningEnd.toLocaleTimeString("en-US", {timeStyle: "short"})}`);

  $('#temp-actual').text(`Temperature: ${convertKtoF(main.temp).toFixed(0)} °F`);
  $('#weather-description').text(`Weather: ${capitalizeFirstLetter(weather.description)}`);
}

async function geolocate() {
  let address = $('#geolocation').val().trim();
  if (address.length == 0) {
    address = "Bellingham, WA"; // default
  }

  let url = "https://api.geoapify.com/v1/geocode/search?text=" + encodeURIComponent(address) + "&apiKey=4e915bbd12764b5191159f65efbf4f47";
  respJson = await (await fetch(url)).json();
  
  const loc = respJson.features[0].properties;

  updateLocation(loc.lat, loc.lon);
}

function convertKtoF(kelvin) {
  return 1.8*(kelvin-273) + 32;
}

function capitalizeFirstLetter(string) {
  return string.charAt(0).toUpperCase() + string.slice(1);
}

function processKey(e)
{
    if (null == e)
        e = window.event ;
    if (e.keyCode == 13)  {
        $('#getgeolocation').trigger('click');
        return false;
    }
}

function onMapClick(e) {
  updateLocation(e.latlng.lat, e.latlng.lng);
}

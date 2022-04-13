window.onload = function () {
  var gallery = document.getElementById("gallery");

  console.log(fetch("https://fleetwood.photos/images/photos.json")
    .then(response => {
      return response.json();
    }));

};
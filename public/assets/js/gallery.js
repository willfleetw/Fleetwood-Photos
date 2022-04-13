$(loadGallery());

async function loadGallery() {
  var photos = await $.get('images/photos.json');
  var gallery = $('#gallery > .row');
  console.log(gallery);

  photos.forEach(photo => {
    var frame = $('<div>', {
      'class': 'col-4 col-6-medium col-12-small'
    });
    
    var link = $('<a>', {
      href: 'images/originals/' + photo['name'],
      'class': 'image fit',
      'data-lightbox': 'mygallery'
    });
    
    var hasTitle = 'title' in photo;
    var hasDesc = 'description' in photo;
    if (hasTitle || hasDesc) {
      var comment = (hasTitle ? photo['title'] : '') 
          + (hasTitle && hasDesc ? ' - ' : '') 
          + (hasDesc ? photo['description'] : '');
      
      link.attr('data-title', comment);
    }
    
    var thumbnail = $('<img>', {
      src: 'images/thumbnails/' + photo['name']
    });

    link.append(thumbnail);
    frame.append(link);

    gallery.append(frame);
  });
}
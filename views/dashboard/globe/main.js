var viewerDiv = document.getElementById('viewerDiv');

var position = {
  coord: new itowns.Coordinates('EPSG:4326', 2.351323, 48.856712),
  range: 25000000
}
var view = new itowns.GlobeView(viewerDiv, position);

itowns.Fetcher.json('http://www.itowns-project.org/itowns/examples/layers/JSONLayers/Ortho.json').then(ortho => {
  var orthoSource = new itowns.WMTSSource(ortho.source);
  var orthoLayer = new itowns.ColorLayer('Ortho', {source: orthoSource});
  view.addLayer(orthoLayer);
});

itowns.Fetcher.json('http://www.itowns-project.org/itowns/examples/layers/JSONLayers/IGN_MNT.json').then(mnt => {
  var mntSource = new itowns.WMTSSource(mnt.source);
  var mntLayer = new itowns.ElevationLayer('IGN_MNT', {source: mntSource});
  view.addLayer(mntLayer);
});

const atmosphere = view.getLayerById('atmosphere');
atmosphere.setRealisticOn(view);

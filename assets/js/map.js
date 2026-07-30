document.addEventListener('alpine:init', () => {
    Alpine.data('mapsidebar', () => ({
        map: null,
        place: '',
        bearing: '',
        distance: 0,
        state: '',
        details: '',
        isLoading: false,
        init() {
            this.map = L.map('map', {
                zoomControl: false // Disable default zoom control
            }).setView([45, -116], 13);

            let self = this;
            // Add zoom control to top-right instead of default top-left
            L.control.zoom({
                position: 'topright'
            }).addTo(self.map);

            L.tileLayer('https://basemap.nationalmap.gov/arcgis/rest/services/USGSTopo/MapServer/tile/{z}/{y}/{x}', {
                maxZoom: 19,
                attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
            }).addTo(self.map);

            L.Control.DetailsView = L.Control.extend({
              onAdd: function(m) {
                  return L.DomUtil.get("widget");
              },
              onRemove: function(m) {

              }
            });
            L.control.DetailsView = function(opts) {
                return new L.Control.DetailsView(opts);
            }

            const d = L.control.DetailsView({position: 'bottomleft'}).addTo(self.map);

            // Add in a crosshair for the map
            var crosshairIcon = L.icon({
                iconUrl: '/assets/img/crosshair.gif',
                iconSize:     [20, 20], // size of the icon
                iconAnchor:   [10, 10], // point of the icon which will correspond to marker's location
            });
            crosshair = new L.marker(self.map.getCenter(), {icon: crosshairIcon, clickable:false});
            crosshair.addTo(self.map);

            navigator.geolocation.getCurrentPosition((pos) => {
                const crd = pos.coords;

                self.map.setView([crd.latitude, crd.longitude], 15);

            }, (err)=> {
                console.warn(`ERROR(${err.code}): ${err.message}`);
            }, this.options);

            // Move the crosshair to the center of the map when the user pans
            this.map.on('move', function(e) {
                let center = self.map.getCenter();
                crosshair.setLatLng(center);
            });

            // Move the crosshair to the center of the map when the user pans
            this.map.on('moveend', function(e) {
                let center = self.map.getCenter();

                let params = {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/x-www-form-urlencoded"
                    },
                    body: new URLSearchParams({ lat: center.lat, lng: center.lng })
                }
                fetch("/geocode", params)
                    .then(r=>r.json())
                    .then(r=>{
                        self.place = r.place;
                        self.state = r.state;
                        self.bearing = r.bearing_direction;
                        self.distance = r.distance;
                    })
            });
        },
        formatplace(){
            let roundedDistance = 0;
            if (this.bearing && this.distance) {
                roundedDistance = Math.round(this.distance);
                if (roundedDistance > 0) {
                    return `<strong>Current Location</strong>: ${roundedDistance} miles ${this.bearing} of ${this.place}, ${this.state}`
                }
            }
            return `<strong>Current Location</strong>: ${this.place}, ${this.state}`
        },
        modalheader(){
            let roundedDistance = 0;
            if (this.bearing && this.distance) {
                roundedDistance = Math.round(this.distance);
                if (roundedDistance > 0) {
                    return `${roundedDistance} miles ${this.bearing} of ${this.place}, ${this.state}`
                }
            }
            return `${this.place}, ${this.state}`
        },
        detailsButtonText() {
            if (this.bearing && this.distance) {
                return `Show Details near ${this.place}`;
            }
            return `Show Details in ${this.place}`;
        },
        hasplace(){
            return !!this.place;
        },
        openModal(el) {
            let center = this.map.getCenter();
            this.isLoading = true;
            let params = {
                method: "POST",
                headers: {
                    "Content-Type": "application/x-www-form-urlencoded"
                },
                body: new URLSearchParams({ lat: center.lat, lng: center.lng })
            }
            let self = this;
            fetch("/details?view=lite", params)
                .then(r=>r.text())
                .then(r=>{
                    self.isLoading = false;
                    self.details = r;

                    const $target = document.querySelector('.modal');
                    $target.style.setProperty("--bulma-modal-z", 1000);

                    $target.classList.add('is-active');
                })


        },
        closeModal(el) {
            const $target = document.querySelector('.modal');
            $target.classList.remove('is-active');
        }
    }));


    // Add a click event on various child elements to close the parent modal
    (document.querySelectorAll('.modal-background, .modal-close, .modal-card-head .delete, .modal-card-foot .button') || []).forEach(($close) => {
        const $target = $close.closest('.modal');

        $close.addEventListener('click', () => {
            const $target = document.querySelector('.modal');
            $target.classList.remove('is-active');
        });
    });

    // Add a keyboard event to close all modals
    document.addEventListener('keydown', (event) => {
        if(event.key === "Escape") {
            const $target = document.querySelector('.modal');
            $target.classList.remove('is-active');
        }
    });

});



document.addEventListener('alpine:init', () => {
    Alpine.data('locate', () => ({
        latitude: '',
        longitude: '',
        searchFilter: '',
        mapboxToken: '',
        automaticCoordinates: false,
        options: {
            enableHighAccuracy: true,
            timeout: 5000,
            maximumAge: 0,
        },
        init() {
            this.mapboxToken = document.getElementById("mapbox_token").value
            navigator.geolocation.getCurrentPosition((pos) => {
                const crd = pos.coords;

                this.latitude = crd.latitude;
                this.longitude = crd.longitude;
                let latInput = document.getElementById('latitude');
                latInput.value = this.latitude;
                let lngInput = document.getElementById('longitude');
                lngInput.value = this.longitude;
                this.automaticCoordinates = true;
            }, (err)=> {
                console.warn(`ERROR(${err.code}): ${err.message}`);
            }, this.options);
        },
        hascoords(){
            return this.latitude && this.longitude
        },
        displaycoords(){
            return `Current Location: ${this.latitude}, ${this.longitude}`
        },
        search(e) {
            let self = this;
            let latInput = document.getElementById('latitude');
            let lngInput = document.getElementById('longitude');
            let formLocate = document.getElementById('locationForm');

            if (latInput.value && lngInput.value && !this.searchFilter){
                formLocate.submit();
            }
            e.preventDefault();

            fetch(`https://api.mapbox.com/search/geocode/v6/forward?q=${this.searchFilter}&access_token=${this.mapboxToken}&autocomplete=false&country=us`)
                .then(r=>r.json())
                .then(r=>{
                    if (r.features && r.features.length > 0) {
                        const feature = r.features[0];
                        if (feature.geometry && feature.geometry.coordinates) {
                            [self.longitude, self.latitude] = feature.geometry.coordinates;
                            latInput.value = self.latitude;
                            lngInput.value = self.longitude;
                            formLocate.submit();
                        }

                    }
                });
        },
    }))
});

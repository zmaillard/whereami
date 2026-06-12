document.addEventListener('alpine:init', () => {
    Alpine.data('locate', () => ({
        latitude: '',
        longitude: '',
        searchFilter: '',
        mapboxToken: '',
        place: '',
        abbreviatedPlace: '',
        automaticCoordinates: false,
        showSearch: false,
        isCurrentLocation: true,
        options: {
            enableHighAccuracy: true,
            timeout: 5000,
            maximumAge: 0,
        },
        init() {
            let self = this;
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


                let params = {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/x-www-form-urlencoded"
                    },
                    body: new URLSearchParams({ lat: this.latitude, lng: this.longitude })
                }
                fetch("/geocode", params)
                    .then(r=>r.json())
                    .then(r=>{
                        self.isCurrentLocation = true;
                        self.place = self.getPlace(r);
                        self.abbreviatedPlace = self.getAbbreviatedPlace(r);
                    })

            }, (err)=> {
                console.warn(`ERROR(${err.code}): ${err.message}`);
                self.showSearch = true;
                self.isCurrentLocation = false;
            }, this.options);
        },
        getAbbreviatedPlace(reverseGeocode) {
            if (reverseGeocode) {
                return reverseGeocode.place
            }
            return ''
        },
        getPlace(reverseGeocode) {
            if (reverseGeocode) {
                return `${reverseGeocode.place}, ${reverseGeocode.state}`
            }
            return ''
        },
        toggle() {
            this.showSearch = !this.showSearch;
        },
        hascoords(){
            return this.latitude && this.longitude
        },
        hasplace(){
            return !!this.place;

        },
        formatplace(){
            return `<strong>${this.isCurrentLocation ? "Current Location" : "User Defined Location"}</strong>: ${this.place}`
        },
        displaycoords(){
            return `<strong>Current Location</strong>: ${this.latitude}, ${this.longitude}`
        },
        detailsButtonText() {
            return `Show Details for ${this.abbreviatedPlace}`
        },
        search(e) {
            let self = this;
            let latInput = document.getElementById('latitude');
            let lngInput = document.getElementById('longitude');

            e.preventDefault();

            fetch(`https://api.mapbox.com/search/geocode/v6/forward?q=${this.searchFilter}&access_token=${this.mapboxToken}&autocomplete=false&country=us`)
                .then(r=>r.json())
                .then(r=>{
                    if (r.features && r.features.length > 0) {
                        const feature = r.features[0];
                        console.log(feature)
                        if (feature.geometry && feature.geometry.coordinates) {
                            [self.longitude, self.latitude] = feature.geometry.coordinates;
                            latInput.value = self.latitude;
                            lngInput.value = self.longitude;

                            self.place = feature.properties.full_address;
                            self.abbreviatedPlace = feature.properties.name;
                            self.isCurrentLocation = false;
                        }

                    }
                });
        },
    }))
});

document.addEventListener('alpine:init', () => {
    Alpine.data('locate', () => ({
        latitude: '',
        longitude: '',
        searchFilter: '',
        automaticCoordinates: false,
        options: {
            enableHighAccuracy: true,
            timeout: 5000,
            maximumAge: 0,
        },
        init() {
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
    }))
});

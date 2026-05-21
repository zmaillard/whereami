document.addEventListener('alpine:init', () => {
    Alpine.data('result', () => ({
        latitude: '',
        longitude: '',
        weather_service_office: '',
        sunrise: '',
        sunset: '',
        elevation: '',
        nearest_summit: '',
        nearest_summit_elevation: '',
        nearest_summit_distance: '',
        tributaries: [],
        ready: false,
        init() {
            let scope = this;
            this.latitude = document.getElementById('latitude').innerText;
            this.longitude = document.getElementById('longitude').innerText;


            fetch(`/query`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/x-www-form-urlencoded",
                },
                body: new URLSearchParams({ lat: this.latitude, lng: this.longitude })})
                .then(response => response.json())
                .then(locate => {
                    scope.weather_service_office = locate.weather_service_office;
                    scope.sunrise = locate.sunrise;
                    scope.sunset = locate.sunset;
                    scope.elevation = locate.elevation;
                    scope.nearest_summit = locate.nearest_summit;
                    scope.nearest_summit_elevation = locate.nearest_summit_elevation;
                    scope.nearest_summit_distance = locate.nearest_summit_distance;
                    scope.tributaries = locate.tributaries;
                    scope.ready = true;
                });
        },
        formattedElevation() {
            return `${this.elevation.toFixed(2)} feet`;
        },
        formattedOfficeLink() {
            return `https://www.weather.gov/${this.weather_service_office}`;
        },
        formattedSummit() {
            return `${this.nearest_summit} (${this.nearest_summit_elevation.toFixed(2)} feet - ${this.nearest_summit_distance.toFixed(2)} km away)`;
        },
        formattedTributaries() {
            return this.tributaries.join(' -> ');
        },
        sink() {
            if (this.tributaries && this.tributaries.length > 0)
                return this.tributaries[this.tributaries.length - 1];
            else
                return '';
        }
    }))
});

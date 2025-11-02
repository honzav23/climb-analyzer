import { Component, AfterViewInit, ViewChild, ElementRef, input } from '@angular/core';
import type { PointCoordinates, Climb } from "../types/AnalysisResponse";
import * as L from 'leaflet'
import 'leaflet-arrowheads'
import {LatLngExpression} from "leaflet";
// import markerIconRed from '../img/marker-icon-red.png';
// import markerIconGreen from '../img/marker-icon-green.png';

@Component({
  selector: 'map-wrapper',
  imports: [],
  templateUrl: './map-wrapper.html',
  styleUrl: './map-wrapper.css'
})
export class MapWrapper implements AfterViewInit {
    map!: L.Map;
    @ViewChild('map', { static: true }) mapContainer!: ElementRef<HTMLDivElement>;
    tripCoords = input<PointCoordinates[]>([]);
    climbs = input<Climb[]>([]);

    ngAfterViewInit() {
        setTimeout(() => {
            this.initMap()
        })
    }

    initMap() {
        const leafletLine: LatLngExpression[] = this.tripCoords().map(coords => [coords.latitude, coords.longitude]);
        this.map = L.map(this.mapContainer.nativeElement).setView(leafletLine[0], 12)

        const tiles = L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            maxZoom: 18,
            minZoom: 3,
            attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
        });

        tiles.addTo(this.map);
        L.polyline(leafletLine, { color: '#FF7F00' }).arrowheads({ frequency: 20, color: "black", yawn: 70, fill: true, size: '12px', weight: 1 }).addTo(this.map);
        for (const climb of this.climbs()) {
            const climbLine: LatLngExpression[] = climb.climbCoordinates.map(coords => [coords.latitude, coords.longitude]);

            // Add the climb start marker
            L.marker(climbLine[0], {
                icon: L.icon({
                    
                    iconUrl: 'assets/marker-icon-green.png',
                    iconSize: [16, 30],
                    iconAnchor: [6, 30],
                })
            }).addTo(this.map);

            // Add the climb end marker
            L.marker(climbLine[climbLine.length - 1], {
                icon: L.icon({
                    iconUrl: 'assets/marker-icon-red.png',
                    iconSize: [16, 30],
                    iconAnchor: [6, 30],
                })
            }).addTo(this.map);
            L.polyline(climbLine, { dashArray: '5, 5' }).addTo(this.map);
        }


        // Make sure that the map properly resizes after it is rendered in DOM
        const observer = new IntersectionObserver((entries) => {
            if (entries[0].isIntersecting) {
                this.map.invalidateSize()
            }
        })
        observer.observe(this.mapContainer.nativeElement)

    }
}

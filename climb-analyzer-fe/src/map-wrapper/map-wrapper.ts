import { Component, AfterViewInit, ViewChild, ElementRef, input } from '@angular/core';
import type { PointCoordinates } from "../types/AnalysisResponse";
import * as L from 'leaflet'
import 'leaflet-arrowheads'
import {LatLngExpression} from "leaflet";

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


        // Make sure that the map properly resizes after it is rendered in DOM
        const observer = new IntersectionObserver((entries) => {
            if (entries[0].isIntersecting) {
                this.map.invalidateSize()
                L.polyline(leafletLine).arrowheads({ frequency: 20, color: "#FFA500", fill: true, size: '12px' }).addTo(this.map);
            }
        })
        observer.observe(this.mapContainer.nativeElement)

    }
}

import {Component, input, OnChanges, OnDestroy, OnInit, SimpleChanges, AfterViewInit, ViewChild, ElementRef} from '@angular/core';
import { type Climb } from "../types/AnalysisResponse";
import Chart from "chart.js/auto";

@Component({
  selector: 'climb-graph',
  imports: [],
  templateUrl: './climb-graph.html',
  styleUrl: './climb-graph.css'
})
export class ClimbGraph implements OnDestroy, AfterViewInit {
    chart: any;
    climbData = input<Climb | null>(null);
    @ViewChild('climbGraph') chartCanvas!: ElementRef<HTMLCanvasElement>;


    // ngOnInit() {
    //     this.createChart()
    // }

    ngAfterViewInit() {
        setTimeout(() => {
            this.createChart()
        })
    }

    // ngOnChanges(changes: SimpleChanges) {
    //     this.createChart()
    // }

    ngOnDestroy() {
        this.chart.destroy();
    }

    getBackgroundColorBasedOnGradient(gradient: number): string {
        if (gradient < 3) {
            return "#5CB85C"
        }
        else if (gradient < 7) {
            return "#FFD700"
        }
        else if (gradient < 10) {
            return "#ff8000"
        }
        return "#e32822"
    }

    // The minimal elevation doesn't have to be the first element of the climb
    getMinimalElevation(climb: Climb): number {
        let currentMinimalElevation = Infinity;
        for (const segment of climb.climbSegments) {
            for (const elevationProf of segment.elevationProfile) {
                if (elevationProf.elevation < currentMinimalElevation) {
                    currentMinimalElevation = elevationProf.elevation;
                }
            }
        }

        return Math.floor(currentMinimalElevation / 100) * 100
    }

    createChart() {
        const ctx = this.chartCanvas.nativeElement.getContext('2d');

        if (!ctx) {
            return
        }

        if (this.chart) {
            this.chart.destroy();
        }
        const climb = this.climbData()
        if (climb === null) {
            return;
        }
        const climbDatasets = []

        for (const segment of climb.climbSegments) {
            const segmentPoints = segment.elevationProfile.map((e) => ({
                x: e.distance,
                y: e.elevation,
            }))
            climbDatasets.push({
                data: segmentPoints,
                pointRadius: 0,
                fill: true,
                backgroundColor: this.getBackgroundColorBasedOnGradient(segment.averageGradient),
                averageGradient: segment.averageGradient.toFixed(1),
                segmentLength: Math.round(segment.segmentLength)
            })
        }

        const gradientLegendItems = [
            { label: '0% - 3%', color: '#5CB85C' },
            { label: '3% - 7%', color: '#FFD700' },
            { label: '7% - 10% ', color: '#ff8000' },
            { label: '10%+', color: '#e32822' }
        ];

        const chartAreaBorder = {
            id: 'chartAreaBorder',
            beforeDraw(chart: Chart) {
                const {ctx, chartArea: {left, top, width, height}} = chart;
                ctx.save();
                ctx.strokeStyle = 'lightgray';
                ctx.lineWidth = 1;
                ctx.strokeRect(left, top, width, height);
                ctx.restore();
            }
        };

        this.chart = new Chart(ctx, {
            type: 'line',
            data: {
                datasets: climbDatasets
            },
            plugins: [chartAreaBorder],
            options: {
                animation: false,
                interaction: {
                    mode: 'dataset',
                    axis: "x",
                    intersect: false,
                },
                elements: {
                    point: {
                        hoverRadius: 0,
                    }
                },
                plugins: {
                    legend: {
                        display: true,
                        labels: {
                            generateLabels: (chart) => {
                                return gradientLegendItems.map((item, index) => ({
                                    datasetIndex: index,
                                    text: item.label,
                                    fillStyle: item.color,
                                    strokeStyle: item.color,
                                    lineWidth: 1,
                                    hidden: false,
                                }));
                            }
                        }
                    },
                    tooltip: {
                        mode: 'dataset',
                        axis: 'x',
                        intersect: false,
                        callbacks: {
                            label: (ctx) => {
                                return [`Length: ${(ctx.dataset as any).segmentLength} m`,
                                    `Average Gradient: ${(ctx.dataset as any).averageGradient}%`]
                            }
                        },
                        filter: (item) => item.dataIndex === 0
                    }
                },
                responsive: true,
                aspectRatio: 16 / 9,
                scales: {
                    y: {
                        title: {
                            display: true,
                            text: 'Elevation (m)',
                            font: {weight: 'bold', size: 27},
                        },
                        grid: {
                            display: false
                        },
                        min: this.getMinimalElevation(climb),
                        ticks: {
                            font: {
                                size: 20
                            },
                            stepSize: 100,
                            callback: function (value, index, ticks) {
                                return value + " m"
                            }
                        }
                    },
                    x: {
                        type: 'linear',
                        title: {
                            display: true,
                            text: 'Distance (km)',
                            font: {weight: 'bold', size: 27},
                        },
                        grid: {
                            display: true,
                            color: 'transparent',
                            drawTicks: true,
                            tickLength: 15,
                            tickColor: 'black'
                        },
                        ticks: {
                            stepSize: 1,
                            font: {
                                size: 20
                            },
                            callback: v => (Number(v) / 1000).toFixed(1) + ' km'
                        },
                        afterBuildTicks: (axis) => {
                            const min = 0
                            const max = climb.length

                            const quarter = climb.length * 0.25;
                            const half = climb.length * 0.5;
                            const threeQuarter = climb.length * 0.75;

                            // create ticks as objects (value + optional label)
                            axis.ticks = [
                                { value: Number(min.toFixed(1)) },
                                { value: Number(quarter.toFixed(1)) },
                                { value: Number(half.toFixed(1)) },
                                { value: Number(threeQuarter.toFixed(1)) },
                                { value: Number(max.toFixed(1)) }
                            ];
                        }
                    }
                }
            }

        });
    }
}

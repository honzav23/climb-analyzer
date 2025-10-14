import {Component, OnInit, input, OnChanges, SimpleChanges, OnDestroy} from '@angular/core';
import Chart from 'chart.js/auto';
import type {ElevationProfilePlotData} from '../types/AnalysisResponse';

@Component({
  selector: 'elevation-graph',
  imports: [],
  templateUrl: './elevation-graph.html',
  styleUrl: './elevation-graph.css'
})
export class ElevationGraph implements OnInit, OnChanges, OnDestroy {
  chart: any;
  elevationProfile = input<ElevationProfilePlotData[]>([])

  ngOnInit() {
      this.createChart()
  }

  ngOnChanges(changes: SimpleChanges) {
      this.createChart()
  }

  ngOnDestroy() {
      this.chart.destroy();
  }

  createChart() {
      if (this.chart) {
        this.chart.destroy();
      }
      const distances = this.elevationProfile().map(e => e.distance)
      const elevations = this.elevationProfile().map(e => e.elevation)

    // Round the lowest elevation of the trip to the nearest 100 meters (for example 388 -> 300)
    const minElevation = Math.floor(Math.min(...elevations) / 100) * 100

    const crosshairPlugin = {
      id: 'elevationCrosshair',
      // The afterDraw hook is used to draw the line after everything else is rendered
      afterDraw: (chart: any) => {
        // Check if a tooltip is currently active (i.e., user is hovering)
        if (chart.tooltip._active && chart.tooltip._active.length > 0) {
          const activeElement = chart.tooltip._active[0];
          const x = activeElement.element.x; // Get the X coordinate of the hovered point
          const yAxis = chart.scales.y;
          const ctx = chart.ctx;

          // Draw the vertical line
          ctx.save();
          ctx.beginPath();
          ctx.lineWidth = 1.5; // Line thickness
          ctx.strokeStyle = '#0a0a0a'; // Orange color for visibility

          // Draw the line from the top of the Y-axis to the bottom
          ctx.moveTo(x, yAxis.top);
          ctx.lineTo(x, yAxis.bottom);
          ctx.stroke();

          ctx.restore();
        }
      }
    };

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

    this.chart = new Chart("elevationGraph", {
      type: 'line',

      data: {// values on X-Axis
        labels: distances,
        datasets: [
          {
            pointRadius: 0,
            fill: true,
            label: "Elevation",
            data: elevations,
            backgroundColor: '#50b012'
          }
        ]
      },
        plugins: [crosshairPlugin, chartAreaBorder],
      options: {
        animation: false,
        plugins: {
          legend: {display: false},
          tooltip: {
            mode: 'index',
            intersect: false,
            axis: 'x',
            callbacks: {
              title: (context) => {
                return context.map((c) => {
                  return (Number(c.label) / 1000).toFixed(1) + " km"
                })
              },
              label: (context) => context.formattedValue + " m"
            }
          }
        },
        responsive: true,
        aspectRatio: 16 / 9,
        scales: {
          y: {
            grid: {
              display: false
            },
            min: minElevation,
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
              min: 0,
              max: distances[distances.length - 1],
              grid: {
                  display: true,
                  color: 'transparent',
                  drawTicks: true,
                  tickLength: 15,
                  tickColor: 'black'
              },
            ticks: {
              font: {
                size: 20
              },
              callback: function (value, index, ticks) {
                const totalPoints = this.ticks.length;
                const indicesToShow = [
                  0, // Start
                  Math.floor(totalPoints / 4), // Quarter
                  Math.floor(totalPoints / 2), // Half
                  Math.floor(totalPoints * 0.75), // Three-quarters
                  totalPoints - 1
                ];

                if (index === 0) {
                  return "0"
                }
                if (indicesToShow.includes(index)) {
                  let val = Number(this.getLabelForValue(index)) / 1000
                  return val.toFixed(1) + ' km';
                }
                return null
              }
            }
          }
        }
      }

    });
  }

}

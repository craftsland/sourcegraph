import React, { useCallback, useState, useMemo } from 'react'
import H from 'history'
import { LineChartContent, BarChartContent, ChartContent, PieChartContent } from 'sourcegraph'
import {
    LineChart,
    ResponsiveContainer,
    CartesianGrid,
    XAxis,
    YAxis,
    Tooltip,
    Legend,
    Line,
    Dot,
    LabelFormatter,
    Bar,
    Cell,
    Rectangle,
    PieChart,
    Pie,
    BarChart,
    Sector,
    ContentRenderer,
    PieLabelRenderProps,
} from 'recharts'
import { createLinkClickHandler } from '../../../shared/src/components/linkClickHandler'
import { niceTicks } from './niceTicks'

const MaybeLink: React.FunctionComponent<React.AnchorHTMLAttributes<unknown>> = ({ children, ...props }) =>
    props.href ? <a {...props}>{children}</a> : (children as React.ReactElement)

const animationDuration = 600

const strokeWidth = 2
const dotRadius = 4
const activeDotRadius = 5

const dateTickFormat = new Intl.DateTimeFormat(undefined, { month: 'long', day: 'numeric' })
const dateTickFormatter = (timestamp: number): string => dateTickFormat.format(timestamp)

// const tooltipLabelFormat = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' })
const tooltipLabelFormat = new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
const tooltipLabelFormatter = (date: number): string => tooltipLabelFormat.format(date)

const toLocaleString = (value: number): string => value.toLocaleString()

export const CartesianChartViewContent: React.FunctionComponent<{
    content: LineChartContent<any, string> | BarChartContent<any, string>
    history: H.History
}> = ({ content, history }) => {
    const linkHandler = useCallback(createLinkClickHandler(history), [history])
    const series: typeof content.series[number][] = content.series
    const ticks = useMemo(() => {
        const allValues = series.flatMap(series => content.data.map(datum => datum[series.dataKey]))
        return niceTicks(Math.min(...allValues), Math.max(...allValues))
    }, [content.data, series])
    const ChartComponent = content.chart === 'line' ? LineChart : BarChart
    return (
        <ResponsiveContainer width="100%" height={12 * 16}>
            <ChartComponent
                className="cartesian-chart-view-content"
                data={content.data}
                margin={{ top: 5, bottom: 5, left: -20, right: 10 }}
            >
                <CartesianGrid vertical={false} />
                <XAxis
                    dataKey={content.xAxis.dataKey}
                    scale={content.xAxis.scale === 'time' ? 'time' : undefined}
                    type={content.xAxis.type}
                    domain={['dataMin', 'dataMax']}
                    tickFormatter={content.xAxis.scale === 'time' ? dateTickFormatter : undefined}
                />
                <YAxis
                    domain={[Math.min(ticks[0], 0), ticks[ticks.length - 1]]}
                    ticks={ticks}
                    type="number"
                    axisLine={false}
                    tickLine={false}
                    tickFormatter={toLocaleString}
                />
                <Tooltip
                    isAnimationActive={false}
                    labelFormatter={tooltipLabelFormatter as LabelFormatter}
                    allowEscapeViewBox={{ x: true, y: true }}
                />
                {series.length > 1 && <Legend />}
                {content.chart === 'line'
                    ? content.series.map(series => (
                          <Line
                              key={series.dataKey as string}
                              name={series.name}
                              dataKey={series.dataKey as string}
                              stroke={series.stroke}
                              strokeWidth={strokeWidth}
                              label={false}
                              animationDuration={animationDuration}
                              animationEasing="ease-in"
                              // eslint-disable-next-line react/jsx-no-bind
                              dot={(props: any) => (
                                  <MaybeLink
                                      href={series.linkURLs?.[props.index]}
                                      key={props.key}
                                      onClick={linkHandler}
                                      className="d-block p-1"
                                  >
                                      <Dot {...props} r={dotRadius} strokeWidth={strokeWidth} />
                                  </MaybeLink>
                              )}
                              // eslint-disable-next-line react/jsx-no-bind
                              activeDot={({ key, ...props }: any) => (
                                  <MaybeLink
                                      href={series.linkURLs?.[props.index]}
                                      key={key}
                                      onClick={linkHandler}
                                      className="d-block p-1"
                                  >
                                      <Dot {...props} r={activeDotRadius} strokeWidth={strokeWidth} />
                                  </MaybeLink>
                              )}
                          />
                      ))
                    : content.series.map(series => (
                          <Bar
                              key={series.dataKey as string}
                              name={series.name}
                              dataKey={series.dataKey as string}
                              fill={series.fill}
                              label={false}
                              // eslint-disable-next-line react/jsx-no-bind
                              shape={({ key, ...props }: any) => (
                                  <MaybeLink href={series.linkURLs?.[props.index]} key={key} onClick={linkHandler}>
                                      <Rectangle {...props} />
                                  </MaybeLink>
                              )}
                          />
                      ))}
            </ChartComponent>
        </ResponsiveContainer>
    )
}

const percentageLabel: ContentRenderer<PieLabelRenderProps> = ({ x, y, ...props }) =>
    props.name + (props.percent ? ': ' + (props.percent * 100).toFixed(0) + '%' : '')

export const PieChartViewContent: React.FunctionComponent<{ content: PieChartContent<any>; history: H.History }> = ({
    content,
    history,
}) => {
    const linkHandler = useCallback(createLinkClickHandler(history), [history])
    const [activeIndex, setActiveIndex] = useState<number>()
    const onMouseEnter = useCallback((data, index) => setActiveIndex(index), [])
    return (
        <ResponsiveContainer height={15 * 16} className="pie-chart-view-content">
            <PieChart>
                {content.pies.map(({ fillKey, dataKey, nameKey, linkURLKey, data }) => (
                    <Pie
                        key={dataKey as string}
                        animationDuration={animationDuration}
                        data={data}
                        dataKey={dataKey as string}
                        nameKey={nameKey as string}
                        outerRadius="70%"
                        // Ensure first sector is at 12 o'clock
                        startAngle={90}
                        endAngle={360 + 90}
                        label={percentageLabel}
                        labelLine={{ stroke: 'var(--body-color)' }}
                        activeIndex={activeIndex}
                        // eslint-disable-next-line react/jsx-no-bind
                        activeShape={({ key, ...props }) => (
                            <MaybeLink key={key} href={linkURLKey && props.payload[linkURLKey]} onClick={linkHandler}>
                                <Sector {...props} />
                            </MaybeLink>
                        )}
                        onMouseEnter={onMouseEnter}
                    >
                        {fillKey && data.map(entry => <Cell key={entry[nameKey]} fill={entry[fillKey]} />)}
                    </Pie>
                ))}
            </PieChart>
        </ResponsiveContainer>
    )
}

export const ChartViewContent: React.FunctionComponent<{ content: ChartContent; history: H.History }> = ({
    content,
    ...props
}) => (
    <>
        {content.chart === 'line' || content.chart === 'bar' ? (
            <CartesianChartViewContent {...props} content={content} />
        ) : content.chart === 'pie' ? (
            <PieChartViewContent {...props} content={content} />
        ) : null}
    </>
)

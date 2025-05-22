const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const TerserPlugin = require('terser-webpack-plugin');
const CompressionPlugin = require('compression-webpack-plugin');
const RemoveEmptyScriptsPlugin = require('webpack-remove-empty-scripts');
const WebpackPwaManifest = require('webpack-pwa-manifest');

const path = require('path');

module.exports = (env, argv) => {
    const isProduction = argv.mode === 'production';

    return {
        entry: {
            admin: ['./src/admin.js', './src/admin.scss'],
        },
        output: {
            path: path.resolve(__dirname, '../public'),
            filename: '[name].js',
            chunkFilename: '[name].chunk.js',
            publicPath: '/'
        },
        module: {
            rules: [
                {
                    test: /\.js$/,
                    loader: 'babel-loader',
                    exclude: /node_modules/,
                },
                {
                    test: /\.scss$/,
                    use: [
                        MiniCssExtractPlugin.loader,
                        {
                            loader: 'css-loader',
                            options: {
                                sourceMap: !isProduction,
                                importLoaders: 2
                            }
                        },
                        {
                            loader: 'sass-loader',
                            options: {
                                sourceMap: !isProduction
                            }
                        }
                    ]
                },
                {
                    test: /\.(woff|woff2|eot|ttf|otf)$/i,
                    type: 'asset/resource',
                    generator: {
                        filename: 'fonts/[name][ext]'
                    }
                },
                {
                    test: /\.(png|svg|jpg|jpeg|gif)$/i,
                    type: 'asset',
                    parser: {
                        dataUrlCondition: {
                            maxSize: 8 * 1024 // 8kb
                        }
                    },
                    generator: {
                        filename: 'images/[name][ext]'
                    }
                }
            ]
        },
        plugins: [
            new RemoveEmptyScriptsPlugin(),
            new MiniCssExtractPlugin({
                filename: 'css/[name].css',
                chunkFilename: 'css/[name].chunk.css',
            }),
            new WebpackPwaManifest({
                filename: 'manifest.json',
                name: 'VGN',
                short_name: 'VGN',
                start_url: '.',
                display: 'standalone',
                theme_color: "#000000",
                background_color: '#ffffff',
                description: 'Viral Game Network manifest',
                orientation: "portrait",
                scope: "/",
                lang: "en-US",
                icons: [
                    {
                        src: path.resolve('./../public/images/vgn.webp'),
                        sizes: [96, 128, 192, 256, 384, 512], // multiple image sizes
                        destination: path.join('images')
                    },
                    {
                        src: path.resolve('./../public/images/vgn.webp'),
                        size: '1024x1024',                // generates one 1024×1024 icon
                        purpose: 'maskable',
                        destination: path.join('images')
                    }
                ]
            }),
            ...(isProduction ? [
                new CompressionPlugin({
                    test: /\.(js|css|html|svg)$/,
                    algorithm: 'gzip',
                })
            ] : []),
        ],
        resolve: {
            alias: {
                // Add aliases here
            },
            extensions: ['.js', '.scss']
        },
        optimization: {
            moduleIds: 'deterministic',
            runtimeChunk: 'single',
            splitChunks: {
                chunks: 'all',
                maxInitialRequests: 5,
                minSize: 20000,
                cacheGroups: {
                    vendors: {
                        test: /[\\/]node_modules[\\/]/,
                        name: 'vendors',
                        priority: -10,
                        chunks: 'all'
                    },
                    common: {
                        name: 'common',
                        minChunks: 2,
                        priority: -20,
                        chunks: 'all',
                        reuseExistingChunk: true
                    }
                }
            },
            minimize: isProduction,
            minimizer: [
                new TerserPlugin({
                    terserOptions: {
                        format: {
                            comments: false,
                        },
                        compress: {
                            drop_console: isProduction,
                        },
                    },
                    extractComments: false,
                }),
            ],
        },
        performance: {
            hints: isProduction ? 'warning' : false,
            maxEntrypointSize: 512000,
            maxAssetSize: 512000
        },
        devtool: isProduction ? 'source-map' : 'eval-source-map',
    };
};

const MiniCssExtractPlugin = require('mini-css-extract-plugin');

const path = require('path');

module.exports = {
    entry: {
        admin: './src/admin.js',
        style: {
            import: './src/style.scss'
        },
    },
    output: {
        path: path.resolve(__dirname, '../public'),
        filename: '[name].js'
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
                    'css-loader',
                    'sass-loader'
                ]
            }
        ]
    },
    plugins: [
        new MiniCssExtractPlugin({
            filename: 'css/[name].css', // Output CSS file
        }),
    ],
    resolve: {
        alias: {
           
        },
    },
};

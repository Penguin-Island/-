const path = require('path');
const CopyPlugin = require('copy-webpack-plugin');
const HtmlPlugin = require('html-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');

const env = process.env.NODE_ENV || 'development';

module.exports = {
    mode: env,
    devtool: env === 'development' ? 'source-map' : false,
    entry: {
        index: './fe/index.ts',
        register: './fe/register.ts',
        friends: './fe/friends.ts',
        game: './fe/game.ts',
        finish: './fe/finish.ts',
    },
    output: {
        path: path.join(__dirname, 'dist'),
        filename: 'js/[name]-[hash].js',
        clean: true,
    },
    module: {
        rules: [
            {
                test: /\.ts$/,
                use: 'ts-loader',
            },
            {
                test: /\.(scss|css)$/,
                use: [MiniCssExtractPlugin.loader, 'css-loader', 'sass-loader'],
            },
        ],
    },
    resolve: {
        modules: ['node_modules'],
        extensions: ['.ts', '.js'],
    },
    plugins: [
        new MiniCssExtractPlugin({
            filename: '[name]-[hash].css',
        }),
        new HtmlPlugin({
            template: 'fe/index.html',
            filename: 'index.html',
            chunks: ['index'],
        }),
        new HtmlPlugin({
            template: 'fe/register.html',
            filename: 'register/index.html',
            chunks: ['register'],
        }),
        new HtmlPlugin({
            template: 'fe/friends.html',
            filename: 'friends/index.html',
            chunks: ['friends'],
        }),
        new HtmlPlugin({
            template: 'fe/game.html',
            filename: 'game/index.html',
            chunks: ['game'],
        }),
        new HtmlPlugin({
            template: 'fe/finish.html',
            filename: 'finish/index.html',
            chunks: ['finish'],
        }),
        // new CopyPlugin({
        //     patterns: [
        //         {from: 'src/index.html', to: path.join(__dirname, 'dist')}
        //     ]
        // })
    ],
};

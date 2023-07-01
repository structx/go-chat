module.exports = {
    mount: {
        public: '/',
        src: '/dist',
    },
    plugins: ['@snowpack/plugin-react-refresh'],
    packageOptions: {
        knownEntrypoints: ['react', 'react-dom'],
    },
};
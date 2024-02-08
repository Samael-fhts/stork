const config = {
    stories: ['../src/**/*.stories.mdx', '../src/**/*.stories.@(js|jsx|ts|tsx)'],

    addons: [
        '@storybook/addon-controls',
        '@storybook/addon-links',
        '@storybook/addon-interactions',
        '@storybook/addon-actions',
        'storybook-addon-mock',
    ],

    framework: {
        name: '@storybook/angular',
        options: {},
    },

    docs: {
        autodocs: false,
    },

    core: {
        disableTelemetry: true,
    }
}

export default config

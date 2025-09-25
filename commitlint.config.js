module.exports = {
    extends: ['@commitlint/config-conventional'],
    rules: {
        'type-enum': [
            2,
            'always',
            [
                'build',     // build system/dependencies
                'chore',     // maintenance
                'ci',        // CI/CD pipeline changes
                'docs',      // documentation
                'feat',      // new feature
                'fix',       // bug fix
                'perf',      // performance improvements
                'refactor',  // code restructuring
                'revert',    // revert commit
                'style',     // formatting / whitespace
                'test'       // add/update tests
            ]
        ],
        'type-empty': [2, 'never'],
        'subject-empty': [2, 'never'],
        'subject-case': [2, 'always', 'lower-case'],
        'subject-full-stop': [2, 'never', '.'],
        'header-max-length': [2, 'always', 50],


        'body-leading-blank': [1, 'always'],
        'body-max-line-length': [2, 'always', 72],
        'footer-leading-blank': [1, 'always'],
    },
};

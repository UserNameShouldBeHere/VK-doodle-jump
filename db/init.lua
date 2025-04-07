datetime = require('datetime')

box.cfg{listen = 3301}
box.schema.user.passwd('pass')

box.schema.space.create('leagues')

box.space.leagues:format({
    {name = 'id', type = 'unsigned'},
    {name = 'name', type = 'string'},
    {name = 'up_cnt', type = 'unsigned'},
    {name = 'stay_cnt', type = 'unsigned'},
})

box.space.leagues:create_index('primary', {type = 'tree', parts = {'id'}})

box.space.leagues:insert{1, 'league1', 3, 4}
box.space.leagues:insert{2, 'league2', 2, 3}
box.space.leagues:insert{3, 'league3', 1, 2}
box.space.leagues:insert{4, 'league4', 0, 2}

-- ===================================

box.schema.space.create('users')

box.space.users:format({
    {name = 'name', type = 'string'},
    {name = 'league', type = 'unsigned', foreign_key = {space = 'leagues', field = 'id'}},
    {name = 'max_score', type = 'unsigned'},
    {name = 'last_update', type = 'datetime'},
})

box.space.users:create_index('name', {type = 'tree', parts = {'name'}})
box.space.users:create_index('max_score', {type = 'tree', unique = false, parts = {'max_score'}})
box.space.users:create_index('score_update', {type = 'tree', parts = {
    {'max_score', sort_order = 'desc'},
    {'last_update', sort_order = 'asc'}
}})
box.space.users:create_index('league_score_update', {type = 'tree', parts = {
    'league',
    {'max_score', sort_order = 'desc'},
    {'last_update', sort_order = 'asc'}
}})

box.schema.func.drop('top_users', {if_exists = true})
box.schema.func.create('users_top', {
    body = [[
        function(limit)
            local lim = args.limit or 10

            local users = box.space.users.index.score_update:select({}, {limit = lim})
            local res = {}
            
            for _, user in ipairs(users) do
                table.insert(res, box.tuple.new({user.name, user.max_score}))
            end

            return res
        end
    ]]
})

box.schema.func.drop('league_users_top', {if_exists = true})
box.schema.func.create('league_users_top', {
    body = [[
        function(args)
            local leagues = {}

            for _, league in ipairs(box.space.leagues.index.primary:select({}, {iterator = box.index.LT})) do
                table.insert(leagues, box.tuple.new({league.id, league.name}))
            end
            
            local lim = args.limit or 10
            local top_users = {}
            
            for _, league in ipairs(leagues) do
                local leagueUsers = {}
                
                for _, user in ipairs(box.space.users.index.league_score_update:select({league[1]}, {limit = lim})) do
                    table.insert(leagueUsers, box.tuple.new({user.name, user.max_score}))
                end

                table.insert(top_users, box.tuple.new({league[2], leagueUsers}))
            end

            return top_users
        end
    ]]
})

box.schema.func.drop('user_score', {if_exists = true})
box.schema.func.create('user_score', {
    body = [[
        function(args)
            return box.space.users.index.name:select({args.name})[1]['max_score']
        end
    ]]
})

-- ===================================

box.schema.func.drop('tmp', {if_exists = true})
box.schema.func.create('tmp', {
    body = [[
        function(args)
            local leagues = {}

            for _, league in ipairs(box.space.leagues.index.primary:select({})) do
                table.insert(leagues, box.tuple.new({league.id, league.name}))
            end
            
            local lim = args.limit or 10
            local top_users = {}
            
            for _, league in ipairs(leagues) do
                local leagueUsers = {}
                local users = box.space.users.index.league_score_update:select({league[1]}, {limit = lim})
                
                for _, user in ipairs(users) do
                    table.insert(leagueUsers, box.tuple.new({user.name, user.max_score}))
                end

                table.insert(top_users, leagueUsers)
            end

            return {
                leagues = leagues,
                top_users = top_users
            }
        end
    ]]
})
